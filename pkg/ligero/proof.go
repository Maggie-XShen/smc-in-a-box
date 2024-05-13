package ligero

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	merkletree "github.com/wealdtech/go-merkletree"
)

type GlobConstants struct {
	flag_num     []bool
	values_num   [][]int
	values_denom []int
	flag_denom   bool
}

type GlobConstantsCodeTest struct {
	flag_num     []bool
	values_num   [][]int
	values_denom []int
	flag_denom   bool
}

var auth_path_end time.Duration
var gen_hash_end time.Duration
var open_col_end time.Duration
var code_end time.Duration
var quadra_end time.Duration
var linear_end time.Duration

type Proof struct {
	MerkleRoot   []byte         `json:"MerkleRoot"`
	ColumnTest   []OpenedColumn `json:"ColumnTest"`
	CodeTest     []int          `json:"CodeTest"`
	QuadraTest   []int          `json:"QuadraTest"`
	LinearTest   []int          `json:"LinearTest"`
	Shares       Shares         `json:"Shares"`
	Seeds        []int          `json:"Seeds"`
	FST_root     []byte         `json:"FST_root"`
	FST_authpath [][]byte       `json:"FST_authpath"`
}

type Shares struct {
	Index      []int   `json:"Index"`
	Values     [][]int `json:"Values"`
	PartyIndex int     `json:"PartyIndex"`
}

type OpenedColumn struct {
	List         []int    `json:"List"`
	Authpath     [][]byte `json:"Authpath"`
	Index        int      `json:"Col_index"`
	Merkle_nonce int      `json:"Merkle_nonce"`
	Code_mask    int      `json:"Code_mask"`
	Linear_mask  int      `json:"Linear_mask"`
	Quadra_mask  int      `json:"Quadra_mask"`
}

func newProof(root []byte, column_check []OpenedColumn, q_code []int, q_quadra []int, q_linear []int, shares Shares, seeds []int, fst_root []byte, fst_authpath [][]byte) *Proof {
	return &Proof{
		MerkleRoot:   root,
		ColumnTest:   column_check,
		CodeTest:     q_code,
		QuadraTest:   q_quadra,
		LinearTest:   q_linear,
		Shares:       shares,
		Seeds:        seeds,
		FST_root:     fst_root,
		FST_authpath: fst_authpath,
	}
}

func (zk *LigeroZK) VerifyProof(proof Proof) (bool, error) {
	//verify fst auth path
	auth_path_start := time.Now()
	fstAuthPathTest, err := zk.verify_fst_authpath(proof.Shares, proof.Seeds, proof.FST_authpath, proof.FST_root)
	if !fstAuthPathTest {
		return false, err
	}
	auth_path_end = time.Since(auth_path_start)

	gen_hash_start := time.Now()
	h1 := zk.generate_hash([][]byte{proof.MerkleRoot})
	//h2 := zk.generate_hash([][]byte{h1, proof.FST_root, ConvertToByteArray(proof.CodeTest), ConvertToByteArray(proof.QuadraTest), ConvertToByteArray(proof.LinearTest)})
	//r4 := zk.generate_random_vector(h2, zk.n_open_col, zk.n_encode)
	gen_hash_end = time.Since(gen_hash_start)

	//verify opened columns are correct
	open_col_start := time.Now()
	openenColumnTest, err := zk.verify_opened_columns(proof.ColumnTest, proof.MerkleRoot)
	if !openenColumnTest {
		return false, err
	}
	open_col_end = time.Since(open_col_start)

	len1 := zk.m * (1 + zk.n_shares)
	len2 := zk.m
	len3 := zk.m

	random_vector := RandVector(h1, len1+len2+len3, zk.q)

	//verify code test proof
	code_start := time.Now()
	r1 := random_vector[:len1]
	codeTest, err := zk.verify_code_proof(proof.CodeTest, r1, proof.ColumnTest)
	if !codeTest {
		return false, err
	}
	code_end = time.Since(code_start)

	//verify quadratic test proof
	quadra_start := time.Now()
	r2 := random_vector[len1 : len1+len2]
	quadraticTest, err := zk.verify_quadratic_constraints(proof.QuadraTest, r2, proof.ColumnTest)
	if !quadraticTest {
		return false, err
	}
	quadra_end = time.Since(quadra_start)

	//verify linear test proof
	linear_start := time.Now()
	r3 := random_vector[len1+len2 : len1+len2+zk.m]
	linearTest, err := zk.verify_linear_proof(proof.Shares, proof.Seeds, proof.LinearTest, r3, proof.ColumnTest)
	if !linearTest {
		return false, err
	}
	linear_end = time.Since(linear_start)

	/**
	type step struct {
		name     string
		time     time.Duration
		duration string
	}

	list := []step{{name: "auth_path", time: auth_path_end, duration: auth_path_end.String()}, {name: "gen_hash", time: gen_hash_end, duration: gen_hash_end.String()}, {name: "open_col", time: open_col_end, duration: open_col_end.String()}, {name: "code_end", time: code_end, duration: code_end.String()}, {name: "quadra_hash", time: quadra_end, duration: quadra_end.String()}, {name: "linear_end", time: linear_end, duration: linear_end.String()}}

	sort.Slice(list, func(i, j int) bool {
		return list[i].time > list[j].time
	})
	fmt.Printf("%+v\n", list)**/

	return true, nil
}

func (zk *LigeroZK) verify_fst_authpath(shares Shares, seeds []int, authpath [][]byte, root []byte) (bool, error) {
	if len(authpath) == 0 || len(root) == 0 {
		return false, fmt.Errorf("fst authpaty or root cannot be empty")
	}

	l2 := len(shares.Values)
	l3 := len(shares.Index)

	list := make([]string, l2*l3+l3)
	index := 0
	for j := 0; j < l2; j++ {
		for n := 0; n < l3; n++ {
			list[index] = fmt.Sprintf("%064b", shares.Values[j][n])
			index++
		}
	}

	for m := 0; m < l3; m++ {
		list[index] = fmt.Sprintf("%064b", seeds[shares.Index[m]])
		index++
	}
	concat := strings.Join(list, "")

	var proof merkletree.Proof
	proof.Hashes = authpath
	proof.Index = uint64(shares.PartyIndex)

	verified, err := merkletree.VerifyProof([]byte(concat), &proof, root)
	if err != nil {
		return false, err
	}

	if !verified {
		return false, fmt.Errorf("failed to verify fst auth path")
	}

	return true, nil

}

func (zk *LigeroZK) verify_opened_columns(open_cols []OpenedColumn, root []byte) (bool, error) {
	if len(open_cols) == 0 || len(root) == 0 {
		return false, fmt.Errorf("opened columns or root cannot be empty")
	}

	for _, col := range open_cols {
		list := make([]int, len(col.List)+1)
		list = append(list, col.List...)
		list = append(list, col.Merkle_nonce)
		concatenated, err := ConvertColumnToString(list)

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
		x := col.Index + 1
		result1, err := zk.Interpolate_at_Point_Code_Test(x_sample, q_code, x, zk.q)
		if err != nil {
			return false, fmt.Errorf("code test failed: x_samples and y_samples length are different")
		}

		result2, err := MulList(randomness, col.List, zk.q)
		if err != nil {
			return false, fmt.Errorf("code test failed: inputs length are different so that multiplication cannot be done")
		}
		result2 = mod(result2+col.Code_mask, zk.q)

		if result1 != result2 {
			return false, fmt.Errorf("code test failed: failed to evaluate the opened column")
		}
	}
	return true, nil
}

func (zk *LigeroZK) verify_quadratic_constraints(q_quadra []int, randomness []int, open_cols []OpenedColumn) (bool, error) {
	//generate x coordicates
	length := len(q_quadra)
	x_sample := make([]int, length)
	for i := 0; i < length; i++ {
		x_sample[i] = i + 1
	}
	for j := 0; j < zk.l; j++ {
		x := mod(-j-1, zk.q)
		result, err := zk.Interpolate_at_Point(x_sample, q_quadra, x, zk.q)
		if err != nil {
			return false, fmt.Errorf("quadratic test failed: failed to evaluat polynomial")
		}
		if result != 0 {
			return false, fmt.Errorf("quadratic test failed: constraints are not satisfied")
		}
	}

	col_test := zk.check_quadra_with_opened_column(q_quadra, randomness, open_cols)

	if !col_test {
		return false, fmt.Errorf("quadratic test failed: failed to evaluate the opened column")
	}

	return true, nil
}

func (zk *LigeroZK) verify_linear_proof(shares Shares, key []int, q_linear []int, randomness []int, open_cols []OpenedColumn) (bool, error) {

	row_test := zk.check_shares_with_opened_column(shares, key, open_cols)
	if !row_test {
		return false, fmt.Errorf("linear test failed: failed to evaluate shares with the opened columns")
	}

	// generate x coordicates
	length := len(q_linear)
	x_sample := make([]int, length)
	for i := 0; i < length; i++ {
		x_sample[i] = i + 1
	}

	for j := 0; j < zk.l; j++ {
		x := mod(-j-1, zk.q)
		result, err := zk.Interpolate_at_Point(x_sample, q_linear, x, zk.q)
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

// func (zk *LigeroZK) check_shares_with_opened_column(parties []rss.Party, key []int, open_cols []OpenedColumn) bool {

// 	size := len(parties)
// 	claims := make([]Claim, size)

// 	for i := 0; i < size; i++ {
// 		sh_list := make([]rss.Share, zk.n_shares)
// 		for j := 0; j < len(parties[0].Shares); j++ {
// 			sh_list[parties[0].Shares[j].Index] = parties[i].Shares[j]
// 		}
// 		claims[i] = Claim{Secret: 0, Shares: sh_list}
// 	}

// 	extended_witness, err := zk.prepare_extended_witness(claims)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	encoded_witness, err := zk.encode_extended_witness(extended_witness, key)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for _, col := range open_cols {
// 		for bl := 0; bl < len(encoded_witness); bl = bl + 1 + zk.n_shares {
// 			for i := 0; i < len(parties[0].Shares); i++ {
// 				rw := bl + parties[0].Shares[i].Index + 1
// 				if (encoded_witness[rw][col.Index]) != col.List[rw] {
// 					return false
// 				}
// 			}
// 		}
// 	}

// 	return true

// }

func (zk *LigeroZK) check_shares_with_opened_column(shares Shares, key []int, open_cols []OpenedColumn) bool {
	size := len(shares.Values)
	claims := make([]Claim, size)

	// Prepare claims for each party concurrently
	var wg sync.WaitGroup
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func(i int) {
			defer wg.Done()
			shList := make([]int, zk.n_shares)
			for j := 0; j < len(shares.Index); j++ {
				shList[shares.Index[j]] = shares.Values[i][j]
			}
			claims[i] = Claim{Secret: 0, Shares: shList}
		}(i)
	}
	wg.Wait()

	// Prepare and encode extended witness concurrently
	extended_witness, err := zk.prepare_extended_witness(claims)
	if err != nil {
		log.Fatal(err)
	}

	encodedWitnesses, err := zk.encode_extended_witness(extended_witness, key)
	if err != nil {
		log.Fatal(err)
	}
	// Check shares with opened columns concurrently
	var result bool
	var mu sync.Mutex
	var wg2 sync.WaitGroup
	wg2.Add(len(open_cols))

	for _, col := range open_cols {
		go func(col OpenedColumn) {
			defer wg2.Done()

			foundMismatch := false
			for bl := 0; bl < len(encodedWitnesses); bl = bl + 1 + zk.n_shares {
				for i := 0; i < len(shares.Index); i++ {
					rw := bl + shares.Index[i] + 1
					if encodedWitnesses[rw][col.Index] != col.List[rw] {
						foundMismatch = true
						break
					}
				}
				if foundMismatch {
					break
				}
			}
			mu.Lock()
			if !foundMismatch {
				result = true
			}
			mu.Unlock()
		}(col)
	}
	wg2.Wait()

	return result
}

// func (zk *LigeroZK) prepare_encode_extended_witness(claims []Claim, key []int) ([][]int, error) {
// 	// Prepare extended witnesses for each claim concurrently
// 	extendedWitnessesChan := make(chan [][]int, len(claims))
// 	errChan := make(chan error, len(claims))
// 	var wg sync.WaitGroup
// 	wg.Add(len(claims))
// 	for _, claim := range claims {
// 		go func(claim Claim) {
// 			defer wg.Done()
// 			extendedWitness, err := zk.prepare_extended_witness([]Claim{claim}) // Pass single claim in a slice
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}
// 			encodedWitness, err := zk.encode_extended_witness(extendedWitness, key)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}
// 			extendedWitnessesChan <- encodedWitness
// 		}(claim)
// 	}
// 	go func() {
// 		wg.Wait()
// 		close(extendedWitnessesChan)
// 	}()

// 	// Collect extended witnesses
// 	extendedWitnesses := make([][]int, len(claims))
// 	for i := range claims {
// 		select {
// 		case extendedWitness := <-extendedWitnessesChan:
// 			extendedWitnesses[i] = extendedWitness
// 		case err := <-errChan:
// 			return nil, err
// 		}
// 	}

// 	return extendedWitnesses, nil
// }

func (zk *LigeroZK) check_quadra_with_opened_column(test_value []int, randomness []int, open_cols []OpenedColumn) bool {
	for _, col := range open_cols {
		result := 0
		index := 0
		for i := 0; i < len(col.List); i = i + zk.n_shares + 1 {
			result += randomness[index] * col.List[i] * (1 - col.List[i])
			index += 1
		}
		result = mod(result+col.Quadra_mask, zk.q)

		if test_value[col.Index] != result {
			return false
		}
	}

	return true
}

func (zk *LigeroZK) check_linear_with_opened_column(test_value []int, randomness []int, open_cols []OpenedColumn) bool {
	for _, col := range open_cols {
		result := 0
		index := 0

		for row := 0; row < len(col.List); row = row + zk.n_shares + 1 {
			temp := col.List[row]
			for j := 1; j < zk.n_shares+1; j++ {
				temp = temp - col.List[row+j]
			}
			result = result + temp*randomness[index]
			index += 1
		}

		result = mod(result+col.Linear_mask, zk.q)

		if test_value[col.Index] != result {
			return false
		}
	}

	return true
}

func (zk *LigeroZK) GetSize(proof Proof) (int64, int64) {
	col_test_size := len(proof.ColumnTest) * (5 + len(proof.ColumnTest[0].List)*8 + len(proof.ColumnTest[0].Authpath)*len(proof.ColumnTest[0].Authpath[0]))
	shares_size := len(proof.Shares.Values)*8*len(proof.Shares.Index) + len(proof.Shares.Index)*8 + 8
	proof_size := (len(proof.CodeTest)+len(proof.QuadraTest)+len(proof.LinearTest)+len(proof.Seeds))*8 + len(proof.MerkleRoot) + len(proof.FST_root) + len(proof.FST_authpath)*len(proof.FST_authpath[0]) + col_test_size
	return int64(proof_size), int64(shares_size)
}
