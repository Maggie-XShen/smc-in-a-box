package rss

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"gonum.org/v1/gonum/stat/combin"
)

type ReplicatedSecretSharing struct {
	n, t, q int
}

type Party struct {
	Index  int     `json:"Index"`
	Shares []Share `json:"Shares"`
}

type Share struct {
	Index int `json:"Index"`
	Value int `json:"Value"`
}

func NewReplicatedSecretSharing(N, T, Q int) (*ReplicatedSecretSharing, error) {
	if T > N {
		return nil, fmt.Errorf("n cannot be less than t")
	}

	//q has to be a prime number
	if !big.NewInt(int64(Q)).ProbablyPrime(0) {
		return nil, fmt.Errorf("q must be a prime number")
	}

	return &ReplicatedSecretSharing{n: N, t: T, q: Q}, nil

}

func (rss *ReplicatedSecretSharing) Split(secret int) ([]Share, []Party, error) {
	//compute total number of shares a secret splits to
	n_sh := combin.Binomial(rss.n, rss.t)

	//compute total number of shares stored by each party
	p_sh := combin.Binomial(rss.n-1, rss.t)

	//generate all shares
	shares := make([]Share, n_sh)
	shares[n_sh-1] = Share{Index: n_sh - 1, Value: secret}
	for i := 0; i < n_sh-1; i++ {
		val, err := rand.Int(rand.Reader, big.NewInt(int64(rss.q)))
		if err != nil {
			return nil, nil, err
		}
		shares[i] = Share{Index: i, Value: int(val.Int64())}
		temp := shares[n_sh-1].Value
		shares[n_sh-1].Value = temp - shares[i].Value
	}
	shares[n_sh-1].Value = mod(shares[n_sh-1].Value, rss.q)

	//generate shares for each party
	list := combin.Combinations(n_sh, p_sh)
	result := make([]Party, rss.n)
	for i := 0; i < rss.n; i++ {
		p_sh := make([]Share, p_sh)
		for j := 0; j < len(list[i]); j++ {
			p_sh[j] = shares[list[i][j]]
		}

		result[i] = Party{Index: i, Shares: p_sh}

	}

	return shares, result, nil

}

func (rss *ReplicatedSecretSharing) Reconstruct(parties []Party) (int, error) {
	//generate a map
	//key: index of the shares the srecret splits to
	//value:a list values associated to the key
	mapping := make(map[int][]int)
	for _, party := range parties {
		for _, sh := range party.Shares {
			mapping[sh.Index] = append(mapping[sh.Index], sh.Value)
		}

	}

	if len(mapping) != combin.Binomial(rss.n, rss.t) {
		return 0, fmt.Errorf("reconstruct failed: missing shares")
	}

	result := 0
	for _, val := range mapping {
		temp, err := findMajority(val)
		if err != nil {
			return 0, err
		}
		result += temp
	}

	result = mod(result, rss.q)

	return result, nil

}

func findMajority(list []int) (int, error) {
	maxCount := 0
	index := -1
	n := len(list)
	for i := 0; i < n; i++ {
		count := 0
		for j := 0; j < n; j++ {
			if list[i] == list[j] {
				count++
			}

		}

		// update maxCount if count of
		// current element is greater
		if count > maxCount {
			maxCount = count
			index = i
		}
	}

	// if maxCount is greater than n/2
	// return the corresponding element
	if maxCount > n/2 { //TODO: replace n with n choose t, minus t
		return list[index], nil
	}

	return 0, fmt.Errorf("reconstruct failed: no majority element")
}

// mod computes a%b and a could be negative number
func mod(a, b int) int {
	return (a%b + b) % b
}
