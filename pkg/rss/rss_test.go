package rss

import (
	"testing"
)

func TestSplitReconstruct(t *testing.T) {
	secret := 1

	rss, err := NewReplicatedSecretSharing(4, 1, 10631)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	shares, parties, err := rss.Split(secret)
	if err != nil {
		t.Fatalf("Split(%v) failed with error %s", secret, err)
	}

	recon, err := rss.Reconstruct(parties)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if recon != secret {
		t.Errorf("shares: %v", shares)
		t.Errorf("partys' shares: %v", parties)
		t.Fatalf("reconstructed secrets do not match original secrets: %v %v", recon, secret)
	}

}
