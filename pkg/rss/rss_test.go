package rss

import (
	"testing"
)

func TestSplitReconstruct(t *testing.T) {
	secret := 7

	rss, err := NewReplicatedSecretSharing(7, 2, 17)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	shares, err := rss.Split(secret)
	if err != nil {
		t.Fatalf("Split(%v) failed with error %s", secret, err)
	}

	recon, err := rss.Reconstruct(shares)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if recon != secret {
		t.Errorf("shares: %v", shares)
		t.Fatalf("reconstructed secrets do not match original secrets: %v %v", recon, secret)
	}

}
