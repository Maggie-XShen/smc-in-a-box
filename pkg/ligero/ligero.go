package ligero

import (
	"fmt"
	"log"
	"math"
	"strings"

	"example.com/SMC/pkg/packed"
	merkletree "github.com/wealdtech/go-merkletree"
)

// n_claim: number of claims
// n_server: number of servers
// m: rows number of rearranged input vecotr in the m*l matrix
// l: columns number of rearranged input vector in the m*l matrix, where n_i = m*l
// t: the maximum number of shares that may be seen without learning anything about the secret;
// use in the secret sharing of each input value
// q: a modulus
// n_encode:the number of shares that each row of rearranged input vector is split into
// n_open_col: number of opened columns

type LigeroZK struct {
	n_claim, m, l, n_server, t, q, n_encode, n_open_col int
}

type EncodedWitness struct {
	matrix [][]int
}

type Proof struct {
	MerkleRoot        []byte         `json:"MerkleRoot"`
	ColumnTest        []OpenedColumn `json:"ColumnTest"`
	CodeTest          []int          `json:"CodeTest"`
	QuadraTest        []int          `json:"QuadraTest"`
	LinearTest        []int          `json:"LinearTest"`
	Code_randomness   []int          `json:"Code_randomness"`
	Quadra_randomness []int          `json:"Quadra_randomness"`
	Linear_randomness []int          `json:"Linear_randomness"`
}

type OpenedColumn struct {
	List        []int    `json:"List"`
	Index       int      `json:"Col_index"`
	Code_mask   int      `json:"Code_mask"`
	Linear_mask int      `json:"Linear_mask"`
	Quadra_mask int      `json:"Quadra_mask"`
	Authpath    [][]byte `json:"Authpath"`
}

type Claim struct {
	Secrets []int
	Shares  []int
}

func NewLigeroZK(N_claim, M, N_server, T, Q, N_open int) (*LigeroZK, error) {
	// m has to larger than 0
	if M <= 0 {
		return nil, fmt.Errorf("m cannot be less than 1")
	}

	if M > N_claim {
		return nil, fmt.Errorf("m cannot be larger than n_input")
	}

	if 3*T+1 > N_server {
		return nil, fmt.Errorf("n_server cannot be less than 3t+1")
	}

	if N_open <= 0 {
		return nil, fmt.Errorf("n_open cannot be less than 1")
	}

	// Calculate l as the upper ceiling of len(slice) divided by m
	L := int(math.Ceil(float64(N_claim) / float64(M)))

	N_encode := 2*N_open + 2*L + 1

	return &LigeroZK{n_claim: N_claim, m: M, l: L, n_server: N_server, t: T, q: Q, n_encode: N_encode, n_open_col: N_open}, nil
}

func (zk *LigeroZK) Generate(claims []Claim) (*Proof, error) {
	if len(claims) == 0 {
		return nil, fmt.Errorf("Invalid input when generating proof: claims are empty")
	}

	/**
	matrix, err := zk.rearrange_input(input, zk.m)
	if err != nil {
		log.Fatal(err)
	}**/

	extended_witness, err := zk.prepare_extended_witness(claims)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("extended_witness: %v\n", extended_witness)

	encoded_witness, err := zk.encode_extended_witness(extended_witness)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("encoded_witness: %v\n", encoded_witness)

	encoded_witeness_columnwise, err := ConvertToColumnwise(encoded_witness)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("encoded_witeness_columnwise: %v\n", encoded_witeness_columnwise)

	seed1 := GenerateRandomness(zk.l, zk.q)
	code_mask := zk.generate_mask(seed1)
	randomness1 := GenerateRandomness(zk.m*(1+zk.n_server), zk.q)

	q_code, err := zk.generate_code_proof(encoded_witness, randomness1, code_mask)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("seed1: %v\n", seed1)
	//fmt.Printf("code_mask: %v\n", code_mask)
	//fmt.Printf("randomness1: %v\n", randomness1)
	//fmt.Printf("q_code: %v\n", q_code)

	seed2 := make([]int, zk.l)
	quadra_mask := zk.generate_mask(seed2)
	randomness2 := GenerateRandomness(zk.m, zk.q)

	q_quadra, err := zk.generate_quadratic_proof(encoded_witness, randomness2, quadra_mask)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("seed2: %v\n", seed2)
	//fmt.Printf("quadra_mask: %v\n", quadra_mask)
	//fmt.Printf("randomness2: %v\n", randomness2)
	//fmt.Printf("q_quadra: %v\n", q_quadra)

	seed3 := make([]int, zk.l)
	linear_mask := zk.generate_mask(seed3)
	linear_rand := GenerateRandomness(zk.m+zk.m*(zk.n_server-zk.t-1), zk.q)
	randomness3 := make([]int, zk.m)
	for i := 0; i < zk.m; i++ {
		randomness3[i] = linear_rand[i]
	}

	q_linear, err := zk.generate_linear_proof(encoded_witness, linear_rand, linear_mask)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("seed3: %v\n", seed3)
	//fmt.Printf("linear_mask: %v\n", linear_mask)
	//fmt.Printf("randomness3: %v\n", randomness3)
	//fmt.Printf("q_linear: %v\n", q_linear)

	//commit to the Extended Witness via Merkle Tree
	tree, leaves, err := zk.generate_merkletree(encoded_witeness_columnwise)
	if err != nil {
		log.Fatal(err)
	}
	//get root of merkletree
	root := tree.Root()

	//generate column check
	randomness0 := GenerateRandomness(zk.n_open_col, len(leaves)) //TODO: need to verify the second parameter

	column_check, err := zk.generate_column_check(tree, leaves, randomness0, code_mask, quadra_mask, linear_mask, encoded_witeness_columnwise)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("opened columns index: %v\n", randomness0)
	//fmt.Printf("column_check: %v\n\n\n", column_check)

	return &Proof{MerkleRoot: root, ColumnTest: column_check, CodeTest: q_code, QuadraTest: q_quadra, LinearTest: q_linear, Code_randomness: randomness1, Quadra_randomness: randomness2, Linear_randomness: randomness3}, nil

}

func (zk *LigeroZK) GetOpenedRows(matrix [][]int, index int) ([][]int, error) {
	if len(matrix) == 0 {
		return nil, fmt.Errorf("Invalid input when getting opened rows: matrix is empty")
	}

	if index < 0 || index >= zk.n_server {
		return nil, fmt.Errorf("Invalid input when getting opened rows: index is not valid")
	}

	result := make([][]int, zk.m)
	pointer := 0
	for i := index; i < len(matrix); i = i + zk.n_server + 1 {
		result[pointer] = matrix[i]
		pointer += 1
	}

	return result, nil

}

func (zk *LigeroZK) Verify(proof Proof) (bool, error) {
	//verify opened columns are correct
	openenColumnTest, err := zk.veify_opened_columns(proof.ColumnTest, proof.MerkleRoot)
	if err != nil {
		return false, err
	}

	if !openenColumnTest {
		return false, fmt.Errorf("openenColumnTest failed")
	}

	//verify code test proof
	codeTest, err := zk.verify_code_proof(proof.CodeTest, proof.Code_randomness, proof.ColumnTest)
	if !codeTest && err != nil {
		return false, err
	}

	//verify quadratic test proof
	quadraticTest, err := zk.verify_quadratic_constraints(proof.QuadraTest, proof.Quadra_randomness, proof.ColumnTest)
	if !quadraticTest && err != nil {
		return false, err
	}

	//verify linear test proof
	linearTest, err := zk.verify_linear_proof(proof.LinearTest, proof.Linear_randomness, proof.ColumnTest)
	if !linearTest && err != nil {
		return false, err
	}

	return true, nil
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
// parameter input: client's input vector
func (zk *LigeroZK) prepare_extended_witness(claims []Claim) ([][]int, error) {
	if len(claims) == 0 {
		return nil, fmt.Errorf("Invalid claims: claims are empty")
	}

	if len(claims[0].Shares) != zk.n_server {
		return nil, fmt.Errorf("Invalid input: Number of shares of each claim must equal to n_server")
	}

	if zk.m > len(claims) {
		return nil, fmt.Errorf("Invalid input: Number of claims must equal or larger than m")
	}

	secrets_num := len(claims[0].Secrets)
	rows := zk.m * (secrets_num + zk.n_server)
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, zk.l)
	}

	index := 0
	for i := 0; i < rows; i = i + secrets_num + zk.n_server {
		for j := 0; j < zk.l; j++ {

			k := 0
			for k < secrets_num {
				matrix[i+k][j] = claims[index].Secrets[k]
				k++
			}
			h := 0
			for h < zk.n_server {
				matrix[i+k+h][j] = claims[index].Shares[h]
				h++
			}
			index++
		}
	}

	return matrix, nil

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

	npss, err := packed.NewPackedSecretSharing(zk.n_encode, zk.n_open_col, zk.l, zk.q)
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
// parameter input:columnwise encoded extended witness,
// each row of the input is a column of eacoded extended witness
func (zk *LigeroZK) generate_merkletree(input [][]int) (*merkletree.MerkleTree, [][]byte, error) {

	// hash each opened column
	leaves := make([][]byte, len(input))

	for i := 0; i < len(input); i++ {
		col := make([]string, len(input))
		for j := 0; j < len(input[0]); j++ {
			col[j] = fmt.Sprintf("%064b", input[i][j])
		}
		//concatenate values in the column to a string
		concatenated := strings.Join(col, "")
		leaves[i] = []byte(concatenated)

	}

	//Create a new Merkle Tree from hashed columns
	tree, err := merkletree.New(leaves)
	if err != nil {
		return nil, nil, err
	}

	return tree, leaves, nil

}

// randomly choose t' columns and get their authentication paths
func (zk *LigeroZK) generate_column_check(tree *merkletree.MerkleTree, leaves [][]byte, cols []int, c_mask []int, q_mask []int, l_mask []int, input [][]int) ([]OpenedColumn, error) {
	column_check := make([]OpenedColumn, len(cols))

	for i := range cols {
		index := cols[i]
		proof, err := tree.GenerateProof(leaves[index])
		if err != nil {
			return nil, err
		}
		column_check[i] = OpenedColumn{List: input[index], Index: index, Code_mask: c_mask[index], Quadra_mask: q_mask[index], Linear_mask: l_mask[index], Authpath: *&proof.Hashes}

	}

	return column_check, nil
}

// generate proof that is used to check if encoded extended witness is encoded correctly
func (zk *LigeroZK) generate_code_proof(input [][]int, randomness []int, mask []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//compute q_code
	r_matrix := make([][]int, 1)
	r_matrix[0] = randomness
	mask_matrix := make([][]int, 1)
	mask_matrix[0] = mask

	temp_matrix, err := MulMatrix(r_matrix, input, zk.q)
	if err != nil {
		return nil, err
	}
	q_code := AddMatrix(temp_matrix, mask_matrix, zk.q)
	if len(q_code) != 1 {
		return nil, fmt.Errorf("Invalid q_code")
	}

	proof := make([]int, zk.n_open_col+zk.l)
	for i := 0; i < zk.n_open_col+zk.l; i++ {
		proof[i] = q_code[0][i]
	}

	return proof, nil

}

// generate proof that is used to check if input is a vector of 0/1
func (zk *LigeroZK) generate_quadratic_proof(input [][]int, randomness []int, mask []int) ([]int, error) {
	//fmt.Printf("input:%v\n", input)
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate q_quadra
	result := make([]int, zk.n_encode)

	index := 0
	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			result[col] += randomness[index] * input[row][col] * (1 - input[row][col])
			//result[col] = mod(result[col], zk.q)
		}
		index += 1
	}

	//fmt.Printf("input:%v\n", result)
	for i := 0; i < len(result); i++ {
		result[i] = result[i] + mask[i]
		result[i] = mod(result[i], zk.q)
	}

	return result, nil

}

// generate proof that is used to check shares of input values are correctly generated
/**
func (zk *LigeroZK) generate_linear_proof(input [][]int, randomness []int, mask []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate lagrange constants
	x_samples := make([]int, zk.t+1)
	for i := 0; i < zk.t+1; i++ {
		x_samples[i] = i + 1
	}

	constants := GenerateLagrangeConstants(x_samples, -1, zk.q)

	//generate q_linear
	result := make([]int, zk.n_encode)

	index := 0
	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			temp := input[row][col]
			for j := 1; j < zk.t+1; j++ {
				temp = temp - constants[j-1]*input[row+j][col]
			}
			result[col] = result[col] + temp*randomness[index]
		}
		index += 1
	}

	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			rand_index := index
			for sh_index := row + 1 + zk.t + 1; sh_index < row+zk.n_server+1; sh_index = sh_index + 1 {
				cons := GenerateLagrangeConstants(x_samples, -sh_index, zk.q)
				temp := input[sh_index][col]
				for j := 1; j < zk.t+1; j++ {
					temp = temp - cons[j-1]*input[row+j][col]
				}
				result[col] = result[col] + temp*randomness[rand_index]
				rand_index += 1
			}

		}
		index += 3
	}

	for i := 0; i < len(result); i++ {
		result[i] = mod(result[i]+mask[i], zk.q)
	}

	return result, nil

}**/

func (zk *LigeroZK) generate_linear_proof(input [][]int, randomness []int, mask []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate lagrange constants
	x_samples := make([]int, zk.n_server)
	for i := 0; i < zk.n_server; i++ {
		x_samples[i] = i + 1
	}

	constants := GenerateLagrangeConstants(x_samples, -1, zk.q)

	//generate q_linear
	result := make([]int, zk.n_encode)

	index := 0
	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			//result[col] = result[col]+input[row][col]
			temp := input[row][col]
			for j := 1; j < zk.n_server+1; j++ {
				temp = temp - constants[j-1]*input[row+j][col]
			}
			result[col] = result[col] + temp*randomness[index]
		}
		index += 1
	}

	for i := 0; i < len(result); i++ {
		result[i] = mod(result[i]+mask[i], zk.q)
	}

	return result, nil

}

func (zk *LigeroZK) veify_opened_columns(open_cols []OpenedColumn, root []byte) (bool, error) {
	if len(open_cols) == 0 || len(root) == 0 {
		return false, fmt.Errorf("opened columns or root cannot be empty")
	}

	for _, col := range open_cols {
		concatenated, err := ConvertColumnToString(col.List)
		if err != nil {
			return false, err
		}

		var proof merkletree.Proof
		proof.Hashes = col.Authpath
		proof.Index = uint64(col.Index)
		verified, err := merkletree.VerifyProof([]byte(concatenated), &proof, root)
		if err != nil {
			return false, err
		}

		if !verified {
			return false, fmt.Errorf("failed to verify the opened column")
		}

	}

	return true, nil

}

func (zk *LigeroZK) verify_code_proof(q_code []int, randomness []int, open_cols []OpenedColumn) (bool, error) {
	//generate x coordicates
	length := len(q_code)
	x_sample := make([]int, length)
	for i := 0; i < length; i++ {
		x_sample[i] = i + 1
	}

	for _, col := range open_cols {
		//fmt.Printf("col:%v\n", col.list)
		//fmt.Printf("index:%d\n", col.col_index)
		//fmt.Printf("randomness:%v\n", randomness)
		x := col.Index + 1
		result1, err := Interpolate_at_Point(x_sample, q_code, x, zk.q)
		if err != nil {
			return false, fmt.Errorf("code test failed: x_samples and y_samples length are different")
		}

		result2, err := MulList(randomness, col.List, zk.q)
		if err != nil {
			return false, fmt.Errorf("code test failed: inputs length are different so that multiplication cannot be done")
		}
		//fmt.Printf("mask:%d\n", col.code_mask)
		//fmt.Printf("result2:%d\n", result2)
		result2 = mod(result2+col.Code_mask, zk.q)

		if result1 != result2 {
			//fmt.Printf("result1:%d\n", result1)
			//fmt.Printf("result2:%d\n", result2)

			return false, fmt.Errorf("code test failed: failed to evaluate the opened column")
		}
	}
	return true, nil
}

func (zk *LigeroZK) verify_quadratic_constraints(q_quadra []int, randomness []int, open_cols []OpenedColumn) (bool, error) {
	//generate x coordicates
	x_sample := make([]int, len(q_quadra))
	for i := 0; i < len(q_quadra); i++ {
		x_sample[i] = i + 1
	}

	for j := 0; j < zk.l; j++ {
		x := mod(-j-1, zk.q)
		result, err := Interpolate_at_Point(x_sample, q_quadra, x, zk.q)
		//fmt.Printf("x_sample:%v\n", x_sample)
		//fmt.Printf("q_quadra:%v\n", q_quadra)
		//fmt.Printf("x:%d\n", x)
		//fmt.Printf("result:%d\n", result)
		if err != nil {
			return false, fmt.Errorf("quadratic test failed: failed to evaluat polynomial")
		}
		if result != 0 {
			return false, fmt.Errorf("quadratic test failed: constraints are not surtisfied")
		}
	}

	col_test := zk.check_quadra_with_opened_column(q_quadra, randomness, open_cols)

	if !col_test {
		return false, fmt.Errorf("quadratic test failed: failed to evaluate the opened column")
	}

	return true, nil
}

func (zk *LigeroZK) verify_linear_proof(q_linear []int, randomness []int, open_cols []OpenedColumn) (bool, error) {
	// generate x coordicates
	x_sample := make([]int, len(q_linear))
	for i := 0; i < len(q_linear); i++ {
		x_sample[i] = i + 1
	}

	for j := 0; j < zk.l; j++ {
		x := mod(-j-1, zk.q)
		result, err := Interpolate_at_Point(x_sample, q_linear, x, zk.q)
		if err != nil {
			return false, fmt.Errorf("linear test failed: failed to evaluat polynomial")
		}
		if result != 0 {
			return false, fmt.Errorf(("linear test failed: shares are not generated correctly"))
		}
	}

	col_test := zk.check_linear_with_opened_column(q_linear, randomness, open_cols)

	if !col_test {
		return false, fmt.Errorf("linear test failed: failed to evaluate the opened column")
	}

	return true, nil
}

func (zk *LigeroZK) check_quadra_with_opened_column(test_value []int, randomness []int, open_cols []OpenedColumn) bool {
	//fmt.Printf("quadra_randomness: %v\n", randomness)
	for _, col := range open_cols {
		//fmt.Printf("col: %v\n", col.list)
		result := 0
		index := 0
		for i := 0; i < len(col.List); i = i + zk.n_server + 1 {
			//fmt.Printf("quadra_randomness: %v\n", randomness[index])
			result += randomness[index] * col.List[i] * (1 - col.List[i])
			index += 1
		}
		//fmt.Printf("result:%d\n", result)
		result = mod(result+col.Quadra_mask, zk.q)
		//fmt.Printf("index:%d\n", col.col_index)
		//fmt.Printf("test_value:%d\n", test_value[col.col_index])
		//fmt.Printf("mask:%d\n", col.quadra_mask)
		//fmt.Printf("result:%d\n", result)

		if test_value[col.Index] != result {
			return false
		}
	}

	return true
}

func (zk *LigeroZK) check_linear_with_opened_column(test_value []int, randomness []int, open_cols []OpenedColumn) bool {

	x_samples := make([]int, zk.n_server)
	for i := 0; i < zk.n_server; i++ {
		x_samples[i] = i + 1
	}

	//generate lagrange constants
	constants := GenerateLagrangeConstants(x_samples, -1, zk.q)

	//fmt.Printf("linear_randomness: %v\n", randomness)
	for _, col := range open_cols {
		//fmt.Printf("col: %v\n", col.list)
		result := 0
		index := 0

		for row := 0; row < len(col.List); row = row + zk.n_server + 1 {
			temp := col.List[row]
			for j := 1; j < zk.n_server+1; j++ {
				temp = temp - constants[j-1]*col.List[row+j]
			}
			result = result + temp*randomness[index]
			index += 1
		}

		result = mod(result+col.Linear_mask, zk.q)

		//fmt.Printf("index:%d\n", col.col_index)
		//fmt.Printf("test_value:%d\n", test_value[col.col_index])
		//fmt.Printf("result:%d\n", result)
		//fmt.Printf("mask:%d\n", col.linear_mask)

		if test_value[col.Index] != result {
			return false
		}
	}

	return true
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
