package ligero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

func ConvertToColumnwise(matrix [][]int) ([][]int, error) {
	if len(matrix) == 0 {
		return nil, fmt.Errorf("matrix cannot be empty")
	}
	result := make([][]int, len(matrix[0]))
	for j := 0; j < len(matrix[0]); j++ {
		temp := make([]int, len(matrix))
		for i := 0; i < len(matrix); i++ {
			temp[i] = matrix[i][j]
		}
		result[j] = temp
	}
	return result, nil
}

func ConvertColumnToString(list []int) (string, error) {
	if len(list) == 0 {
		return "", fmt.Errorf("list cannot be empty")
	}

	col := make([]string, len(list))
	for i := 0; i < len(list); i++ {
		col[i] = fmt.Sprintf("%064b", list[i])
	}
	//concatenate values in the column to a string
	concatenated := strings.Join(col, "")

	return concatenated, nil

}

func AddMatrix(matrix1 [][]int, matrix2 [][]int) [][]int {
	result := make([][]int, len(matrix1))
	for i, a := range matrix1 {
		for j := range a {
			result[i] = append(result[i], matrix1[i][j]+matrix2[i][j])
		}
	}
	return result
}

func SubMatrix(matrix1 [][]int, matrix2 [][]int) [][]int {
	result := make([][]int, len(matrix1))
	for i, a := range matrix1 {
		for j := range a {
			result[i] = append(result[i], matrix1[i][j]-matrix2[i][j])
		}
	}
	return result
}

func MulMatrix(matrix1 [][]int, matrix2 [][]int) [][]int {
	result := make([][]int, len(matrix1))
	for i := 0; i < len(matrix1); i++ {
		result[i] = make([]int, len(matrix1))
		for j := 0; j < len(matrix2); j++ {
			for k := 0; k < len(matrix2); k++ {
				result[i][j] += matrix1[i][k] * matrix2[k][j]
			}
		}
	}
	return result
}

func GenerateRandomness(length int, q int) []int {
	randomness := make([]int, length)
	//rand.Seed(time.Now().UnixNano())
	checkMap := map[int]bool{}
	for i := 0; i < length; i++ {
		for {
			value, err := rand.Int(rand.Reader, big.NewInt(int64(q)))
			if err == nil && !checkMap[int(value.Int64())] {
				checkMap[int(value.Int64())] = true
				randomness[i] = int(value.Int64())
				break
			}

		}
	}

	return randomness
}

// lagrange_constants_for_point returns lagrange constants for the given x
func GenerateLagrangeConstants(x_samples []int, x int, q int) []int {

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
				num = mod(num*(xj-x), q)
				denum = mod(denum*(xj-xi), q)
			}
		}
		constants[i] = mod(num*inverse(denum, q), q)
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
func inverse(a int, q int) int {

	a = (a + q) % q
	b := egcd_binary(a, q)
	return b
}

// mod computes a%b and a could be negative number
func mod(a, b int) int {
	return (a%b + b) % b
}
