// TODO: add reference
package ligero

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/chacha20"
)

var mainCipher *chacha20.Cipher

type CryptoRandSource struct {
}

func NewCryptoRandSource() CryptoRandSource {
	return CryptoRandSource{}
}

func (c CryptoRandSource) Seed(seeds ...interface{}) {
	// Concatenate the seeds
	var concatenatedSeeds []byte
	for _, seed := range seeds {
		switch v := seed.(type) {
		case int:
			seedBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(seedBytes, uint64(v))
			concatenatedSeeds = append(concatenatedSeeds, seedBytes...)
		case string:
			concatenatedSeeds = append(concatenatedSeeds, []byte(v)...)
		case []byte:
			concatenatedSeeds = append(concatenatedSeeds, v...)
		default:
			panic(errors.New("unsupported seed type"))
		}
	}

	key := sha256.Sum256(concatenatedSeeds)

	nonce := make([]byte, 24)

	var err error
	mainCipher, err = chacha20.NewUnauthenticatedCipher(key[:], nonce)
	if err != nil {
		panic(err)
	}
}

func (c CryptoRandSource) Int63(q int64) int64 {
	if mainCipher == nil {
		panic(errors.New("crypto seed not set"))
	}

	//var b [8]byte
	b := make([]byte, 8)
	zeroes := make([]byte, 8)
	mainCipher.XORKeyStream(b, zeroes)

	// mask off sign bit to ensure positive number
	return int64(binary.LittleEndian.Uint64(b[:])&(1<<63-1)) % q
}

func RandVector(seed []byte, length int, q int) []int {
	random_vector := make([]int, length)

	checkMap := map[int]bool{}
	crs := NewCryptoRandSource()
	crs.Seed(seed)
	for i := 0; i < length; i++ {
		for {
			randomNumber := int(crs.Int63(int64(q)))

			if !checkMap[randomNumber] {
				checkMap[randomNumber] = true
				random_vector[i] = randomNumber
				break
			}

		}
	}
	return random_vector
}
