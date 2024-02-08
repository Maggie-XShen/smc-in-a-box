package ligero

import (
	"fmt"
	"testing"
)

func TestRand(t *testing.T) {
	crs := NewCryptoRandSource()
	crs.Seed(123, "abc", 456, []byte("efg"), "ghi")
	for i := 0; i < 5; i++ {
		randomNumber := int(crs.Int63(10631))
		fmt.Printf("Random number %d: %d\n", i+1, randomNumber)
	}

	crs.Seed(789, 987)
	for i := 0; i < 5; i++ {
		randomNumber := int(crs.Int63(10631))
		fmt.Printf("Random number %d: %d\n", i+1, randomNumber)
	}

}
