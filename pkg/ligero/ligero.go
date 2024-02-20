package ligero

import (
	crypto_rand "crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"

	"strings"

	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/rss"
	merkletree "github.com/wealdtech/go-merkletree"
	"golang.org/x/crypto/sha3"
	"gonum.org/v1/gonum/stat/combin"
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
	n_secret, n_shares, m, l, n_server, t, q, n_encode, n_open_col int
}

type Claim struct {
	Secret int
	Shares []rss.Share
}

func NewLigeroZK(N_secret, M, N_server, T, Q, N_open int) (*LigeroZK, error) {
	// m has to larger than 0
	if M <= 0 {
		return nil, fmt.Errorf("m cannot be less than 1")
	}

	if M > N_secret {
		return nil, fmt.Errorf("m cannot be larger than n_secrets")
	}

	if 3*T+1 > N_server {
		return nil, fmt.Errorf("n_server cannot be less than 3t+1")
	}

	if N_open <= 0 {
		return nil, fmt.Errorf("n_open cannot be less than 1")
	}

	//compute total number of shares a secret splits to
	N_shares := combin.Binomial(N_server, T)

	// Calculate l as the upper ceiling of len(slice) divided by m
	L := int(math.Ceil(float64(N_secret) / float64(M)))

	N_encode := 2*N_open + 2*L + 1

	return &LigeroZK{n_secret: N_secret, n_shares: N_shares, m: M, l: L, n_server: N_server, t: T, q: Q, n_encode: N_encode, n_open_col: N_open}, nil
}

func (zk *LigeroZK) GenerateProof(secrets []int) ([]*Proof, error) {
	claims, party_sh, err := zk.preprocess(secrets)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("claims: %v\n", claims)
	//fmt.Printf("party_sh: %v\n", party_sh)

	extended_witness, err := zk.prepare_extended_witness(claims)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("extended_witness: %v\n", extended_witness)

	seed0 := generate_seeds(zk.n_shares+1, zk.q)
	encoded_witness, err := zk.encode_extended_witness(extended_witness, seed0)
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
	root := tree.Root()

	//generate a vector of random numbers using the hash of merkle tree root as seed
	len1 := zk.m * (1 + zk.n_server)
	len2 := zk.m
	len3 := zk.m + zk.m*(zk.n_server-zk.t-1)
	h1 := zk.generate_hash([][]byte{root})
	random_vector := RandVector(h1, len1+len2+len3, zk.q)

	//generate code test
	seed1 := generate_seeds(zk.l, zk.q)
	code_mask := zk.generate_mask(seed1)
	r1 := random_vector[:len1]

	q_code, err := zk.generate_code_proof(encoded_witness, r1, code_mask)
	if err != nil {
		log.Fatal(err)
	}

	//generate quadratic test
	seed2 := make([]int, zk.l)
	quadra_mask := zk.generate_mask(seed2)
	r2 := random_vector[len1 : len1+len2]

	q_quadra, err := zk.generate_quadratic_proof(encoded_witness, r2, quadra_mask)
	if err != nil {
		log.Fatal(err)
	}

	//generate linear test
	seed3 := make([]int, zk.l)
	linear_mask := zk.generate_mask(seed3)
	r3 := random_vector[len1+len2:]

	q_linear, err := zk.generate_linear_proof(encoded_witness, r3, linear_mask)
	if err != nil {
		log.Fatal(err)
	}

	//generate FST root
	fst_tree, fst_leaves, err := zk.generate_fst_merkletree(party_sh, seed0)
	if err != nil {
		log.Fatal(err)
	}
	fst_root := fst_tree.Root()

	h2 := zk.generate_hash([][]byte{h1, fst_root, ConvertToByteArray(q_code), ConvertToByteArray(q_quadra), ConvertToByteArray(q_linear)})

	//generate column check
	r4 := RandVector(h2, zk.n_open_col, len(leaves)) //TODO: need to verify the third parameter
	column_check, err := zk.generate_column_check(tree, leaves, r4, code_mask, quadra_mask, linear_mask, encoded_witeness_columnwise)
	if err != nil {
		log.Fatal(err)
	}

	//generate proof for each party
	proofs := make([]*Proof, zk.n_server)
	for i := 0; i < zk.n_server; i++ {

		fst_proof, err := fst_tree.GenerateProof(fst_leaves[i])
		if err != nil {
			log.Fatal("could not generate fst authentication path")
		}

		proofs[i] = newProof(root, column_check, q_code, q_quadra, q_linear, party_sh[i], seed0, fst_root, fst_proof.Hashes)
	}

	return proofs, nil

}

func (zk *LigeroZK) preprocess(secrets []int) ([]Claim, [][]rss.Party, error) {
	n_secret := len(secrets)
	if n_secret == 0 || n_secret != zk.n_secret {
		return nil, nil, fmt.Errorf("Invalid input when generating proof: wrong number of secrets")
	}

	nrss, err := rss.NewReplicatedSecretSharing(zk.n_server, zk.t, zk.q)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	claims := make([]Claim, n_secret)
	party_sh := make([][]rss.Party, zk.n_server)
	for i := 0; i < zk.n_server; i++ {
		party_sh[i] = make([]rss.Party, n_secret)
	}

	for i := 0; i < n_secret; i++ {
		sh, party, err := nrss.Split(secrets[i])
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		claims[i] = Claim{Secret: secrets[i], Shares: sh}
		for j := 0; j < len(party); j++ {
			party_sh[j][i] = party[j]
		}

	}

	return claims, party_sh, nil
}

// Generate shares of each value in the input vector, store them with input values in a matrix, which is called extended witness
// parameter input: client's input vector
func (zk *LigeroZK) prepare_extended_witness(claims []Claim) ([][]int, error) {
	if len(claims) == 0 {
		return nil, fmt.Errorf("Invalid claims: claims are empty")
	}

	if len(claims[0].Shares) != zk.n_shares {
		return nil, fmt.Errorf("Invalid input: Number of shares of each claim is not correct")
	}

	if zk.m > len(claims) {
		return nil, fmt.Errorf("Invalid input: Number of claims must equal or larger than m")
	}

	secrets_num := 1
	rows := zk.m * (secrets_num + zk.n_shares)
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, zk.l)
	}

	index := 0
	for i := 0; i < rows; i = i + secrets_num + zk.n_shares {
		for j := 0; j < zk.l; j++ {

			k := 0
			for k < secrets_num {
				matrix[i+k][j] = claims[index].Secret
				k++
			}
			h := 0
			for h < zk.n_server {
				matrix[i+k+h][j] = claims[index].Shares[h].Value
				h++
			}
			index++
		}
	}

	return matrix, nil

}

// encode extended witness row-by-row using packed secret sharing
func (zk *LigeroZK) encode_extended_witness(input [][]int, key []int) ([][]int, error) {
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

	crs := NewCryptoRandSource()
	//shamir-secret sharing each row in input
	for i := 0; i < len(input); i++ {
		nonce := i / (1 + zk.n_shares)

		crs.Seed(key[i%(1+zk.n_shares)], nonce)

		shares, err := npss.Split(input[i], int(crs.Int63(int64(zk.q))))
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

func (zk *LigeroZK) generate_fst_merkletree(party_sh [][]rss.Party, seeds []int) (*merkletree.MerkleTree, [][]byte, error) {
	// generate and hash each leaf
	l1 := len(party_sh)
	l2 := len(party_sh[0])
	l3 := len(party_sh[0][0].Shares)

	if l1 == 0 || l2 == 0 || l3 == 0 {
		log.Fatal("party_sh is invalid")
	}
	leaves := make([][]byte, l1)
	for i := 0; i < l1; i++ {
		list := make([]string, l2*l3+l3)
		index := 0
		for j := 0; j < l2; j++ {
			for n := 0; n < l3; n++ {
				list[index] = fmt.Sprintf("%064b", party_sh[i][j].Shares[n].Value)
				index++
			}
		}

		for m := 0; m < l3; m++ {
			list[index] = fmt.Sprintf("%064b", seeds[party_sh[i][0].Shares[m].Index])
			index++
		}
		concat := strings.Join(list, "")
		leaves[i] = []byte(concat)
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
		column_check[i] = OpenedColumn{List: input[index], Index: index, Code_mask: c_mask[index], Quadra_mask: q_mask[index], Linear_mask: l_mask[index], Authpath: proof.Hashes}

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

func (zk *LigeroZK) generate_linear_proof(input [][]int, randomness []int, mask []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Invalid input: Input is empty")
	}

	if len(input) != zk.m*(1+zk.n_server) || len(input[0]) != zk.n_encode {
		return nil, fmt.Errorf("Invalid input")
	}

	//generate q_linear
	result := make([]int, zk.n_encode)

	index := 0
	for row := 0; row < len(input); row = row + zk.n_server + 1 {
		for col := 0; col < len(input[0]); col++ {
			//result[col] = result[col]+input[row][col]
			temp := input[row][col]
			for j := 1; j < zk.n_server+1; j++ {
				temp = temp - input[row+j][col]
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

func (zk *LigeroZK) generate_mask(seeds []int) []int {

	mask := make([]int, zk.n_encode)

	npss, err := packed.NewPackedSecretSharing(zk.n_encode, zk.t, zk.l, zk.q)
	if err != nil {
		log.Fatal(err)
	}

	shares, err := npss.Split(seeds, 1)
	if err != nil {
		log.Fatal(err)
	}

	for j := 0; j < zk.n_encode; j++ {
		mask[j] = shares[j].Value
	}

	return mask
}

func (zk *LigeroZK) generate_hash(input [][]byte) []byte {
	if len(input) == 0 {
		log.Fatal("input of hash function could not be empty")
	}
	size := 0
	for _, d := range input {
		size += len(d)
	}

	concat, i := make([]byte, size), 0
	for _, d := range input {
		i += copy(concat[i:], d)
	}

	hash := sha3.Sum256(concat)

	return hash[:]
}

func generate_seeds(size int, q int) []int {
	seeds := make([]int, size)
	//rand.Seed(time.Now().UnixNano())
	checkMap := map[int]bool{}
	for i := 0; i < size; i++ {
		for {
			value, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(q)))
			if err == nil && !checkMap[int(value.Int64())] {
				checkMap[int(value.Int64())] = true
				seeds[i] = int(value.Int64())
				break
			}

		}
	}

	return seeds
}
