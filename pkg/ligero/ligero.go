package ligero

import (
	"fmt"
	"log"
	"math"
	"strings"

	"example.com/SMC/pkg/packed"
	merkletree "github.com/wealdtech/go-merkletree"
)

// n_i: length of input vector
// n_s: number of servers
// m: rows number of rearranged input vecotr in the m*l matrix
// l: columns number of rearranged input vector in the m*l matrix, where n_i = m*l
// t: number of malicious servers
// q: a modulus
// n_encode:
// n_open_col: number of opened column

type LigeroZK struct {
	n_input, m, l, n_server, t, q, n_encode, n_open_col int
}

type Proof struct {
	MerkleRoot  []byte
	ColumnCheck []OpenedColumn
	Q_code      []int
	Q_quadra    []int
	Q_linear    []int
}

type OpenedColumn struct {
	list     []int
	authpath *merkletree.Proof
}

func NewLigeroZK(Ni, M, Ns, T, Q, N_open int) (*LigeroZK, error) {
	// m has to larger than 0
	if M <= 0 {
		return nil, fmt.Errorf("m cannot be less than 1")
	}

	if M > Ni {
		return nil, fmt.Errorf("m cannot be less than n_s")
	}

	// Calculate l as the upper ceiling of len(slice) divided by m
	L := int(math.Ceil(float64(Ni) / float64(M)))

	NEncode := 3*T + 2*L + 1

	//TODO: check if N_open is less than NEncode and larger than 0

	return &LigeroZK{n_input: Ni, m: M, l: L, n_server: Ns, t: T, q: Q, n_encode: NEncode, n_open_col: N_open}, nil
}

func (zk *LigeroZK) Generate(input []int) (*Proof, error) {

	/**
	matrix, err := zk.rearrange_input(input, zk.m)
	if err != nil {
		log.Fatal(err)
	}**/

	extended_witness, err := zk.prepare_extended_witness(input)
	if err != nil {
		log.Fatal(err)
	}

	encoded_witness, err := zk.encode_extended_witness(extended_witness)
	if err != nil {
		log.Fatal(err)
	}

	encoded_witeness_columnwise, err := ConvertToColumnwise(encoded_witness)
	if err != nil {
		log.Fatal(err)
	}

	//commit to the Extended Witness via Merkle Tree
	tree, leaves, err := zk.generate_merkletree(encoded_witeness_columnwise)
	if err != nil {
		log.Fatal(err)
	}
	//get root of merkletree
	root := tree.Root()

	//generate column check
	randomness0 := GenerateRandomness(zk.n_open_col, len(leaves)) //TODO: need to verify the second parameter
	column_check, err := zk.generate_column_check(tree, leaves, randomness0, encoded_witeness_columnwise)
	if err != nil {
		log.Fatal(err)
	}

	randomness1 := GenerateRandomness(zk.m*(1+zk.n_server), zk.q)
	q_code, err := zk.generate_code_proof(encoded_witness, randomness1)
	if err != nil {
		log.Fatal(err)
	}

	randomness2 := GenerateRandomness(zk.m, zk.q)
	q_quadra, err := zk.generate_quadratic_proof(encoded_witness, randomness2)
	if err != nil {
		log.Fatal(err)
	}

	randomness3 := GenerateRandomness(zk.m, zk.q)
	q_linear, err := zk.generate_linear_proof(encoded_witness, randomness3)
	if err != nil {
		log.Fatal(err)
	}

	return &Proof{MerkleRoot: root, ColumnCheck: column_check, Q_code: q_code, Q_quadra: q_quadra, Q_linear: q_linear}, nil

}

func (zk *LigeroZK) Verify() {

}

/**
// rearrange input vector to matrix
func (zk *LigeroZK) rearrange_input(input []int, m int) ([][]int, error) {
	if m > len(input) {
		return nil, errors.New("Invalid input: Number of elements in the inputs must equal or larger than m")
	}

	// Calculate l as the upper ceiling of len(slice) divided by m
	l := int(math.Ceil(float64(len(input)) / float64(m)))

	matrix := make([][]int, m)
	for i := range matrix {
		matrix[i] = make([]int, l)
	}

	index := 0
	for i := 0; i < m; i++ {
		for j := 0; j < l; j++ {
			matrix[i][j] = input[index]
			index++
			if index >= len(input) {
				break
			}
		}
	}

	return matrix, nil

}**/

// Generate shares of each value in the input vector, store them with input values in a matrix, which is called extended witness
func (zk *LigeroZK) prepare_extended_witness(input []int) ([][]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if zk.m > len(input) {
		return nil, fmt.Errorf("Invalid input: Number of elements in the input must equal or larger than m")
	}

	matrix1 := make([][]int, zk.m*(1+zk.n_server))
	for i := range matrix1 {
		matrix1[i] = make([]int, zk.l)
	}

	matrix2 := make([][]int, zk.m)
	for i := range matrix2 {
		matrix2[i] = make([]int, zk.l)
	}

	npss, err := packed.NewPackedSecretSharing(zk.n_server, zk.t, 1, zk.q)

	if err != nil {
		log.Fatal(err)
	}

	index1 := 0
	for i := 0; i < zk.m*(1+zk.n_server); i = i + zk.n_server + 1 {
		for j := 0; j < zk.l; j++ {
			if index1 < len(input) {
				matrix1[i][j] = input[index1]
				index1++
			}

			//shamir-secret sharing each item in input
			shares, err := npss.Split([]int{matrix1[i][j]})
			if err != nil {
				log.Fatal(err)
			}

			index2 := 0
			for v := i + 1; v < i+1+zk.n_server; v++ {
				matrix1[v][j] = shares[index2].Value
				index2++
			}

		}
	}

	return matrix1, nil

}

// encode extended witness row-by-row using packed secret sharing
func (zk *LigeroZK) encode_extended_witness(input [][]int) ([][]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.l {
		return nil, fmt.Errorf("Invalid input")
	}

	matrix := make([][]int, len(input))
	for i := range matrix {
		matrix[i] = make([]int, zk.n_encode)
	}

	npss, err := packed.NewPackedSecretSharing(zk.n_encode, zk.t, zk.l, zk.q)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(input); i++ {
		//shamir-secret sharing each row in input
		shares, err := npss.Split(input[i])
		if err != nil {
			log.Fatal(err)
		}

		values := make([]int, zk.n_encode)
		for j := 0; j < zk.n_encode; j++ {
			values[j] = shares[j].Value
		}

		matrix[i] = values

	}

	return matrix, nil
}

// commit encoded extended witness via Merkle Tree
func (zk *LigeroZK) generate_merkletree(input [][]int) (*merkletree.MerkleTree, [][]byte, error) {

	// hash each opened column
	leaves := make([][]byte, len(input))

	for j := 0; j < len(input[0]); j++ {
		col := make([]string, len(input))
		for i := 0; i < len(input); i++ {
			col[i] = fmt.Sprintf("%064b", input[i][j])
		}
		//concatenate values in the column to a string
		concatenated := strings.Join(col, "")
		leaves[j] = []byte(concatenated)
	}

	//Create a new Merkle Tree from hashed columns
	tree, err := merkletree.New(leaves)
	if err != nil {
		return nil, nil, err
	}

	return tree, leaves, nil

}

// randomly choose t' columns and get their authentication paths
func (zk *LigeroZK) generate_column_check(tree *merkletree.MerkleTree, leaves [][]byte, col_index []int, input [][]int) ([]OpenedColumn, error) {
	column_check := make([]OpenedColumn, len(col_index))
	//choose random columns
	for i, _ := range col_index {
		index := col_index[i]
		proof, err := tree.GenerateProof(leaves[index])
		if err != nil {
			return nil, err
		}
		column_check[i] = OpenedColumn{list: input[index], authpath: proof}
	}
	return column_check, nil
}

func (zk *LigeroZK) veify_opened_columns(open_cols []OpenedColumn, root []byte) (bool, error) {
	//TODO: chcek cols and roor are not empty
	for _, col := range open_cols {
		concatenated, err := ConvertColumnToString(col.list)
		if err != nil {
			return false, err
		}

		verified, err := merkletree.VerifyProof([]byte(concatenated), col.authpath, root)
		if err != nil {
			return false, err
		}

		if !verified {
			return false, nil
		}

	}

	return true, nil

}

// generate proof that is used to check if encoded extended witness is encoded correctly
func (zk *LigeroZK) generate_code_proof(input [][]int, randomness []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate mask
	r := GenerateRandomness(zk.l, zk.q)
	mask := zk.generate_mask(r)

	//compute q_code
	r_matrix := make([][]int, 1)
	r_matrix[0] = randomness
	mask_matrix := make([][]int, 1)
	mask_matrix[0] = mask

	temp_matrix := MulMatrix(r_matrix, input)
	q_code := AddMatrix(temp_matrix, mask_matrix)
	if len(q_code) != 1 {
		return nil, fmt.Errorf("Invalid q_code")
	}

	return q_code[0], nil

}

// generate proof that is used to check if input is a vector of 0/1
func (zk *LigeroZK) generate_quadratic_proof(input [][]int, randomness []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate mask
	r := make([]int, zk.l)
	mask := zk.generate_mask(r)

	//generate q_quadra
	result := mask

	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			result[col] += randomness[row] * input[row][col] * (1 - input[row][col])
		}
	}

	return result, nil

}

// generate proof that is used to check shares of input values are correctly generated
func (zk *LigeroZK) generate_linear_proof(input [][]int, randomness []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate mask
	r := make([]int, zk.l)
	mask := zk.generate_mask(r)

	//generate lagrange constants
	x_samples := make([]int, zk.n_server)
	for i := 0; i < zk.n_server; i++ {
		x_samples[i] = i + 1
	}

	constants := GenerateLagrangeConstants(x_samples, -1, zk.q)

	//generate q_linear
	result := mask

	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			result[col] = randomness[row] * input[row][col]
		}

		for j := 1; j < zk.n_server+1; j++ {
			for col := 0; col < len(input[0]); col++ {
				result[col] -= constants[j-1] * input[row+j][col]
			}
		}

	}

	return result, nil

}

func (zk *LigeroZK) verify_conrrectness() {

}

func (zk *LigeroZK) verify_code_proof() {

}

func (zk *LigeroZK) verify_quadratic_constraints() {

}

func (zk *LigeroZK) verify_linear_proof() {

}

func (zk *LigeroZK) generate_mask(input []int) []int {
	mask := make([]int, zk.n_encode)

	npss, err := packed.NewPackedSecretSharing(zk.n_encode, zk.t, zk.l, zk.q)
	if err != nil {
		log.Fatal(err)
	}

	shares, err := npss.Split(input)
	if err != nil {
		log.Fatal(err)
	}

	for j := 0; j < zk.n_encode; j++ {
		mask[j] = shares[j].Value
	}

	return mask
}
