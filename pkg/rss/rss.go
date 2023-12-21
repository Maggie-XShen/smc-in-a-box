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

type Share struct {
	PartyIndex int     `json:"PartyIndex"`
	Values     []Value `json:"Values"`
}

type Value struct {
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

func (rss *ReplicatedSecretSharing) Split(secret int) ([]Share, error) {
	//compute total number of values a secret splits to
	n_values := combin.Binomial(rss.n, rss.t)

	//compute total number of values stored by each party
	p_values := combin.Binomial(rss.n-1, rss.t)

	//generate all values
	values := make([]int, n_values)
	values[n_values-1] = secret
	for i := 0; i < n_values-1; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(rss.q)))
		if err != nil {
			return nil, err
		}
		values[i] = int(value.Int64())
		values[n_values-1] -= values[i]
	}
	values[n_values-1] = mod(values[n_values-1], rss.q)

	//generate share for each party
	list := combin.Combinations(n_values, p_values)
	result := make([]Share, rss.n)
	for i := 0; i < rss.n; i++ {
		vals_party := make([]Value, p_values)
		for j := 0; j < len(list[i]); j++ {
			x := Value{Index: list[i][j], Value: values[list[i][j]]}
			vals_party[j] = x
		}

		result[i] = Share{PartyIndex: i + 1, Values: vals_party}

	}

	return result, nil

}

func (rss *ReplicatedSecretSharing) Reconstruct(shares []Share) (int, error) {
	//generate a map
	//key: index of the values the srecret splits to
	//value:a list values associated to the key
	mapping := make(map[int][]int)
	for _, sh := range shares {
		for _, val := range sh.Values {
			mapping[val.Index] = append(mapping[val.Index], val.Value)
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
	if maxCount > n/2 {
		return list[index], nil
	}

	return 0, fmt.Errorf("reconstruct failed: no majority element")
}

// mod computes a%b and a could be negative number
func mod(a, b int) int {
	return (a%b + b) % b
}
