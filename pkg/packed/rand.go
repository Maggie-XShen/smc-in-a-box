package packed

import (
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/chacha20"
)

// A cryptographically secure pseudorandom number generator that we should use
// for everything in order to ensure both good randomness and repeatability.
// It's based on the chacha20 cipher.

type CryptoRandSource struct {
	mainCipher *chacha20.Cipher
}

func NewCryptoRandSource() CryptoRandSource {
	return CryptoRandSource{
		mainCipher: new(chacha20.Cipher),
	}
}

func (c *CryptoRandSource) Int63() int64 {
	if c.mainCipher == nil {
		panic(errors.New("crypto seed not set"))
	}

	//var b [8]byte
	b := make([]byte, 8)
	zeroes := make([]byte, 8)

	c.mainCipher.XORKeyStream(b, zeroes)

	// mask off sign bit to ensure positive number
	return int64(binary.LittleEndian.Uint64(b[:]) & (1<<63 - 1))
}

func (c *CryptoRandSource) Seed(seed int64) {
	var err error
	key := make([]byte, 32)
	seedU := uint64(seed)
	binary.BigEndian.PutUint64(key, seedU)
	nonce := make([]byte, 24)

	c.mainCipher, err = chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		panic(err)
	}
}
