package sqlstore

import (
	"testing"
)

func TestInsertClient(t *testing.T) {
	db := NewDB("test")

	//create a new server for experiment 1
	err := db.InsertServerShare("exp1", "s1", 1, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetSharesPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_shares=%v, want %v", got, want)
			}
		}
	}

	// create same server for experiment 1
	err = db.InsertServerShare("exp1", "s1", 1, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetSharesPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_shares=%v, want %v", got, want)
			}
		}
	}

	// create second client for experiment 1
	err = db.InsertServerShare("exp1", "s2", 1, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		Shares, err := db.GetSharesPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(Shares), 2
			if got != want {
				t.Fatalf("num_shares=%v, want %v", got, want)
			}
		}
	}

	// create new client for experiment 2
	err = db.InsertServerShare("exp2", "s2", 1, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetSharesPerExperiment("exp2")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_shares=%v, want %v", got, want)
			}
		}
	}

}
