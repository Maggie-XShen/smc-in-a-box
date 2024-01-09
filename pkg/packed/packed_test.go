package packed

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestInit_invalid(t *testing.T) {

	if _, err := NewPackedSecretSharing(10, 8, 2, 41); err != nil {
		t.Fatalf("err: %v", err)
	}

	if _, err := NewPackedSecretSharing(5, 1, 2, 41); err != nil {
		t.Fatalf("err: %v", err)
	}

	if _, err := NewPackedSecretSharing(5, 1, 2, 131); err != nil {
		t.Fatalf("err: %v", err)
	}

}

func TestSplit(t *testing.T) {
	secret := []int{10, 25, 35}

	npss, err := NewPackedSecretSharing(20, 8, 3, 41)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if npss.n != 20 {
		t.Fatalf("bad: %v", npss.n)
	}

	if npss.t != 8 {
		t.Fatalf("bad: %v", npss.t)
	}

	if npss.k != 3 {
		t.Fatalf("bad: %v", npss.k)
	}

	if npss.q != 41 {
		t.Fatalf("bad: %v", npss.q)
	}

	seed := 99
	shares, err := npss.Split(secret, seed)
	if err != nil {
		t.Fatalf("Split(%v) failed with error %s", secret, err)
	}

	if len(shares) != 20 {
		t.Fatalf("bad: %v", shares)
	}

}

func TestReconstruct(t *testing.T) {
	secrets := []int{100, 50}

	npss, err := NewPackedSecretSharing(20, 8, 2, 151)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	seed := 99
	shares, err := npss.Split(secrets, seed)
	if err != nil {
		t.Fatalf("Split(%v) failed with error %s", secrets, err)
	}

	// randomly pick t+k shares from total N shares
	checkMap := map[int]bool{}
	parts := make([]Share, npss.t+npss.k)
	for i := 0; i < npss.t+npss.k; i++ {
		for {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(npss.n)))
			if err == nil && !checkMap[int(idx.Int64())] {
				checkMap[int(idx.Int64())] = true
				parts[i] = shares[int(idx.Int64())]
				break
			}
		}
	}

	recon, err := npss.Reconstruct(parts)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if len(recon) != len(secrets) {
		t.Fatalf("reconstructed secrets do not match original secrets: %v %v", recon, secrets)
	}

	for i := 1; i < len(recon); i++ {
		if recon[i] != secrets[i] {
			t.Errorf("parts: %v", parts)
			t.Fatalf("reconstructed secrets do not match original secrets: %v %v", recon, secrets)
		}
	}

}
