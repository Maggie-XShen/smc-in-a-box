package ligero

import (
	"fmt"
	"log"
	"strings"

	"example.com/SMC/pkg/rss"
	merkletree "github.com/wealdtech/go-merkletree"
)

type Proof struct {
	MerkleRoot   []byte         `json:"MerkleRoot"`
	ColumnTest   []OpenedColumn `json:"ColumnTest"`
	CodeTest     []int          `json:"CodeTest"`
	QuadraTest   []int          `json:"QuadraTest"`
	LinearTest   []int          `json:"LinearTest"`
	PartyShares  []rss.Party    `json:"PartyShares"`
	Seeds        []int          `json:"Seeds"`
	FST_root     []byte         `json:"FST_root"`
	FST_authpath [][]byte       `json:"FST_authpath"`
}

type OpenedColumn struct {
	List         []int    `json:"List"`
	Index        int      `json:"Col_index"`
	Merkle_nonce int      `json:"Merkle_nonce"`
	Code_mask    int      `json:"Code_mask"`
	Linear_mask  int      `json:"Linear_mask"`
	Quadra_mask  int      `json:"Quadra_mask"`
	Authpath     [][]byte `json:"Authpath"`
}

func newProof(root []byte, column_check []OpenedColumn, q_code []int, q_quadra []int, q_linear []int, party_sh []rss.Party, seeds []int, fst_root []byte, fst_authpath [][]byte) *Proof {
	return &Proof{
		MerkleRoot:   root,
		ColumnTest:   column_check,
		CodeTest:     q_code,
		QuadraTest:   q_quadra,
		LinearTest:   q_linear,
		PartyShares:  party_sh,
		Seeds:        seeds,
		FST_root:     fst_root,
		FST_authpath: fst_authpath,
	}
}

func (zk *LigeroZK) VerifyProof(proof Proof) (bool, error) {
	//verify fst auth path
	fstAuthPathTest, err := zk.verify_fst_authpath(proof.PartyShares, proof.Seeds, proof.FST_authpath, proof.FST_root)
	if !fstAuthPathTest {
		return false, err
	}

	h1 := zk.generate_hash([][]byte{proof.MerkleRoot})
	//h2 := zk.generate_hash([][]byte{h1, proof.FST_root, ConvertToByteArray(proof.CodeTest), ConvertToByteArray(proof.QuadraTest), ConvertToByteArray(proof.LinearTest)})
	//r4 := zk.generate_random_vector(h2, zk.n_open_col, zk.n_encode)

	//verify opened columns are correct
	openenColumnTest, err := zk.verify_opened_columns(proof.ColumnTest, proof.MerkleRoot)
	if !openenColumnTest {
		return false, err
	}

	len1 := zk.m * (1 + zk.n_server)
	len2 := zk.m
	len3 := zk.m + zk.m*(zk.n_server-zk.t-1)

	random_vector := RandVector(h1, len1+len2+len3, zk.q)

	//verify code test proof
	r1 := random_vector[:len1]
	codeTest, err := zk.verify_code_proof(proof.CodeTest, r1, proof.ColumnTest)
	if !codeTest {
		return false, err
	}

	//verify quadratic test proof
	r2 := random_vector[len1 : len1+len2]
	quadraticTest, err := zk.verify_quadratic_constraints(proof.QuadraTest, r2, proof.ColumnTest)
	if !quadraticTest {
		return false, err
	}

	//verify linear test proof
	r3 := random_vector[len1+len2 : len1+len2+zk.m]
	linearTest, err := zk.verify_linear_proof(proof.PartyShares, proof.Seeds, proof.LinearTest, r3, proof.ColumnTest)
	if !linearTest {
		return false, err
	}

	return true, nil
}

func (zk *LigeroZK) verify_fst_authpath(party_sh []rss.Party, seeds []int, authpath [][]byte, root []byte) (bool, error) {
	if len(authpath) == 0 || len(root) == 0 {
		return false, fmt.Errorf("fst authpaty or root cannot be empty")
	}

	l2 := len(party_sh)
	l3 := len(party_sh[0].Shares)

	list := make([]string, l2*l3+l3)
	index := 0
	for j := 0; j < l2; j++ {
		for n := 0; n < l3; n++ {
			list[index] = fmt.Sprintf("%064b", party_sh[j].Shares[n].Value)
			index++
		}
	}

	for m := 0; m < l3; m++ {
		list[index] = fmt.Sprintf("%064b", seeds[party_sh[0].Shares[m].Index])
		index++
	}
	concat := strings.Join(list, "")

	var proof merkletree.Proof
	proof.Hashes = authpath
	proof.Index = uint64(party_sh[0].Index)

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
		result1, err := Interpolate_at_Point(x_sample, q_code, x, zk.q)
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
	x_sample := make([]int, len(q_quadra))
	for i := 0; i < len(q_quadra); i++ {
		x_sample[i] = i + 1
	}

	for j := 0; j < zk.l; j++ {
		x := mod(-j-1, zk.q)
		result, err := Interpolate_at_Point(x_sample, q_quadra, x, zk.q)
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

func (zk *LigeroZK) verify_linear_proof(parties []rss.Party, key []int, q_linear []int, randomness []int, open_cols []OpenedColumn) (bool, error) {
	row_test := zk.check_shares_with_opened_column(parties, key, open_cols)
	if !row_test {
		return false, fmt.Errorf("linear test failed: failed to evaluate shares with the opened columns")
	}

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

func (zk *LigeroZK) check_shares_with_opened_column(parties []rss.Party, key []int, open_cols []OpenedColumn) bool {

	size := len(parties)
	claims := make([]Claim, size)

	for i := 0; i < size; i++ {
		sh_list := make([]rss.Share, zk.n_shares)
		for j := 0; j < len(parties[0].Shares); j++ {
			sh_list[parties[0].Shares[j].Index] = parties[i].Shares[j]
		}
		claims[i] = Claim{Secret: 0, Shares: sh_list}
	}

	extended_witness, err := zk.prepare_extended_witness(claims)
	if err != nil {
		log.Fatal(err)
	}

	encoded_witness, err := zk.encode_extended_witness(extended_witness, key)
	if err != nil {
		log.Fatal(err)
	}

	for _, col := range open_cols {
		for bl := 0; bl < len(encoded_witness); bl = bl + 1 + zk.n_shares {
			for i := 0; i < len(parties[0].Shares); i++ {
				rw := bl + parties[0].Shares[i].Index + 1
				if (encoded_witness[rw][col.Index]) != col.List[rw] {
					return false
				}
			}
		}
	}

	return true

}

func (zk *LigeroZK) check_quadra_with_opened_column(test_value []int, randomness []int, open_cols []OpenedColumn) bool {
	for _, col := range open_cols {
		result := 0
		index := 0
		for i := 0; i < len(col.List); i = i + zk.n_server + 1 {
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

		for row := 0; row < len(col.List); row = row + zk.n_server + 1 {
			temp := col.List[row]
			for j := 1; j < zk.n_server+1; j++ {
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

func (zk *LigeroZK) GetProofSize(proof Proof) int64 {
	col_test_size := len(proof.ColumnTest) * (5 + len(proof.ColumnTest[0].List)*8 + len(proof.ColumnTest[0].Authpath)*len(proof.ColumnTest[0].Authpath[0]))
	shares_size := len(proof.PartyShares) * (8 + len(proof.PartyShares[0].Shares)*16)
	size := (len(proof.CodeTest)+len(proof.QuadraTest)+len(proof.LinearTest)+len(proof.Seeds))*8 + len(proof.MerkleRoot) + len(proof.FST_root) + len(proof.FST_authpath)*len(proof.FST_authpath[0]) + col_test_size + shares_size
	return int64(size)
}
