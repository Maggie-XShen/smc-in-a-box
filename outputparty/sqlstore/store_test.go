package sqlstore

import (
	"encoding/json"
	"testing"
)

func TestInsertServerShare(t *testing.T) {
	db := NewDB("test")

	//create a new server for experiment 1
	shares, _ := json.Marshal([][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}})
	err := db.InsertServerShare("exp1", "s1", shares)
	if err != nil {
		t.Log(err)
	} else {
		shares, err := db.GetSharesPerExperiment("exp1")
		if err != nil {
			t.Log(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Logf("num_servers=%v, want %v", got, want)
			}
		}
	}

	// create same server for experiment 1
	err = db.InsertServerShare("exp1", "s1", shares)
	if err != nil {
		t.Log(err)
	}

	// create second client for experiment 1
	err = db.InsertServerShare("exp1", "s2", shares)
	if err != nil {
		t.Fatal(err)
	} else {
		Shares, err := db.GetSharesPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(Shares), 2
			if got != want {
				t.Fatalf("num_servers=%v, want %v", got, want)
			}
		}
	}

	// create new client for experiment 2
	err = db.InsertServerShare("exp2", "s2", shares)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetSharesPerExperiment("exp2")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_servers=%v, want %v", got, want)
			}
		}
	}

}
