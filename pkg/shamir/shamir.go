package shamir

import (
	"github.com/hashicorp/vault/shamir"
)

// n is number of shares, t is threshold
func Split(data []byte, n, t int) ([][]byte, error) {
	return shamir.Split(data, n, t)
}

func Reconstruct(shares [][]byte) ([]byte, error) {
	return shamir.Combine(shares)
}
