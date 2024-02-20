package packed

import (
	"fmt"
	"math/big"
	"math/rand"
)

// n: the number of shares that a vector of secrets are split into
// t: the maximum number of shares that may be seen
// without learning anything about the secret
// k: the number of secrets shared together
// q: a modulus
// t + k: the minimum number of shares needed to reconstruct the secret

type PackedSecretSharing struct {
	n, t, k, q int
}

type Share struct {
	Index int `json:"Index"`
	Value int `json:"Value"`
}

func NewPackedSecretSharing(N, T, K, Q int) (*PackedSecretSharing, error) {
	if T+K > N {
		return nil, fmt.Errorf("n cannot be less than t+k")
	}

	//constrainrs on t and k
	if K < 1 {
		return nil, fmt.Errorf("k must be at least 1")
	}

	//q has to be a prime number
	if !big.NewInt(int64(Q)).ProbablyPrime(0) {
		return nil, fmt.Errorf("q must be a prime number")
	}

	return &PackedSecretSharing{n: N, t: T, k: K, q: Q}, nil

}

// Split takes k secrets and generates n shares.
// Each returned share was attached a tag used to reconstruct the secrets.
func (p *PackedSecretSharing) Split(secrets []int, seed int) ([]Share, error) {
	if len(secrets) == 0 {
		return nil, fmt.Errorf("cannot split an empty secret")
	}

	x_samples, y_samples, err := p.sample_packed_polynomial(secrets, seed)

	if err != nil {
		return nil, err
	}

	shares := make([]Share, p.n)
	for idx := range shares {
		xCoordinate := idx + 1
		shares[idx].Index = xCoordinate
		shares[idx].Value = p.interpolate_at_point(x_samples, y_samples, xCoordinate)
	}

	return shares, nil

}

// Reconstruct takes t+k shares and reconstruct k secrets
func (p *PackedSecretSharing) Reconstruct(parts []Share) ([]int, error) {
	//need t+k shares to reconstruct
	if len(parts) < p.t+p.k {
		return nil, fmt.Errorf("cannot reconstruct, as number of shares less than t+k")
	}

	if len(parts) > p.n {
		return nil, fmt.Errorf("cannot reconstruct, as number of shares more than n")
	}

	var x_samples []int
	var y_samples []int
	for i := 0; i < len(parts); i++ {
		x_samples = append(x_samples, parts[i].Index)
		y_samples = append(y_samples, parts[i].Value)
	}

	secrets := make([]int, p.k)
	for i := 0; i < p.k; i++ {
		xCoordinate := mod(-i-1, p.q)
		secrets[i] = p.interpolate_at_point(x_samples, y_samples, xCoordinate)
	}
	return secrets, nil

}

// sample_packed_polynomial constructs a random polynomial of t+k-1 degree
func (p *PackedSecretSharing) sample_packed_polynomial(secrets []int, seed int) ([]int, []int, error) {
	x_samples := make([]int, p.k+p.t)
	for i := 0; i < p.k+p.t; i++ {
		x_samples[i] = mod(-i-1, p.q)
	}

	randomness_values := make([]int, p.t)
	seedValue := int64(seed)
	MainCSRNG = rand.New(NewCryptoRandSource())
	MainCSRNG.Seed(seedValue)
	for i := 0; i < p.t; i++ {
		randomNumber := MainCSRNG.Int63()
		randomness_values[i] = int(randomNumber % int64(p.q))
	}

	/**
	r := rand.New(rand.NewSource(int64(seed)))
	checkMap := map[int]bool{}
	for i := 0; i < p.t; i++ {
		for {
			value := r.Intn(p.q)
			if !checkMap[int(value)] {
				checkMap[int(value)] = true
				randomness_values[i] = int(value)
				break
			}

		}
	}**/

	y_samples := append(secrets, randomness_values...)

	return x_samples, y_samples, nil
}

// interpolate_at_point takes t+k sample points and returns
// the value at a given x using a lagrange interpolation.
func (p *PackedSecretSharing) interpolate_at_point(x_samples []int, y_samples []int, x int) int {

	constants := p.lagrange_constants_for_point(x_samples, x)
	y := 0
	for i := 0; i < len(y_samples); i++ {
		y = y + y_samples[i]*constants[i]
	}
	return mod(y, p.q)
}

// lagrange_constants_for_point returns lagrange constants for the given x
func (p *PackedSecretSharing) lagrange_constants_for_point(x_samples []int, x int) []int {

	constants := make([]int, len(x_samples))
	for i := range constants {
		constants[i] = 0
	}

	for i := 0; i < len(constants); i++ {
		xi := x_samples[i]
		num := 1
		denum := 1
		for j := 0; j < len(constants); j++ {
			if j != i {
				xj := x_samples[j]
				num = mod(num*(xj-x), p.q)
				denum = mod(denum*(xj-xi), p.q)
			}
		}
		constants[i] = mod(num*p.inverse(denum), p.q)
	}

	return constants
}

// from http://www.ucl.ac.uk/~ucahcjm/combopt/ext_gcd_python_programs.pdf
func egcd_binary(a int, b int) int {
	u, v, s, t, r := 1, 0, 0, 1, 0
	for (mod(a, 2) == 0) && (mod(b, 2) == 0) {
		a, b, r = a/2, b/2, r+1
	}

	alpha, beta := a, b

	for mod(a, 2) == 0 {
		a = a / 2
		if (mod(u, 2) == 0) && (mod(v, 2) == 0) {
			u, v = u/2, v/2
		} else {
			u, v = (u+beta)/2, (v-alpha)/2
		}

	}

	for a != b {
		if mod(b, 2) == 0 {
			b = b / 2
			if (mod(s, 2) == 0) && (mod(t, 2) == 0) {
				s, t = s/2, t/2
			} else {
				s, t = (s+beta)/2, (t-alpha)/2
			}
		} else if b < a {
			a, b, u, v, s, t = b, a, s, t, u, v
		} else {

			b, s, t = b-a, s-u, t-v
		}

	}

	return s
}

// inverse calculates the inverse of a number
func (p *PackedSecretSharing) inverse(a int) int {

	a = (a + p.q) % p.q
	b := egcd_binary(a, p.q)
	return b
}

// mod computes a%b and a could be negative number
func mod(a, b int) int {
	return (a%b + b) % b
}
