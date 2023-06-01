package shamir

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
)

func TestSplitAndCombine(t *testing.T) {
	message := "10"

	n := 5
	k := 3

	fmt.Printf("Message:\t%s\n\n", message)

	if k > n {
		fmt.Printf("Cannot do this, as k greater than n")
		os.Exit(0)
	}

	shares, err := Split([]byte(message), n, k)
	if err != nil {
		t.Fatalf("SplitShares(secret, %d, %d) failed with error %s", n, k, err)
	}

	parts := [][]byte{shares[1], shares[3], shares[4]}
	fmt.Printf("shares: %s,%s,%s\n\n", hex.EncodeToString(parts[0]), hex.EncodeToString(parts[1]), hex.EncodeToString(parts[2]))

	reconstructed, err := Reconstruct(parts)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Println("Reconstructed:", string(reconstructed))

	if !bytes.Equal(reconstructed, []byte(message)) {
		t.Errorf("CombineShares(%v) = %v, want %v", parts, reconstructed, message)
	}
}
