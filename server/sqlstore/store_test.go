package sqlstore

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInsertClient(t *testing.T) {
	db := NewDB("test")

	//create a new client for experiment 1
	err := db.InsertClient("exp1", "c1")
	if err != nil {
		t.Fatal(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp1")
		if err != nil {
			t.Log(err)
		} else {
			got, want := len(clients), 1
			if got != want {
				t.Logf("num_clients=%v, want %v", got, want)
			}
		}
	}

	// create same client for experiment 1
	err = db.InsertClient("exp1", "c1")
	if err != nil {
		t.Log(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp1")
		if err != nil {
			t.Log(err)
		} else {
			got, want := len(clients), 1
			if got != want {
				t.Logf("num_clients=%v, want %v", got, want)
			}
		}
	}

	// create second client for experiment 1
	err = db.InsertClient("exp1", "c2")
	if err != nil {
		t.Log(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp1")
		if err != nil {
			t.Log(err)
		} else {
			got, want := len(clients), 2
			if got != want {
				t.Logf("num_clients=%v, want %v", got, want)
			}
		}
	}

	// create new client for experiment 2
	err = db.InsertClient("exp2", "c1")
	if err != nil {
		t.Log(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp2")
		if err != nil {
			t.Log(err)
		} else {
			got, want := len(clients), 1
			if got != want {
				t.Logf("num_clients=%v, want %v", got, want)
			}
		}
	}

	//DeleteDB("test.db")

}

func TestInsertClientShare(t *testing.T) {
	db := NewDB("test")

	shares, _ := json.Marshal([][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}})
	//create a new client's share
	err := db.InsertClientShare("exp1", "c1", shares)
	if err != nil {
		t.Log(err)
	}

	// create client share which already exist
	newShares, _ := json.Marshal([][]int{{0, 0, 0}})
	err = db.InsertClientShare("exp1", "c1", newShares)
	if err != nil {
		t.Log(err)
	}

	record, err := db.GetClientShares("exp1", "c1")
	if err != nil {
		t.Log(err)
	}
	var array2D [][]int
	err = json.Unmarshal([]byte(record.Shares), &array2D)
	if err != nil {
		t.Log(err)
	} else {
		got, want := array2D, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
		if !reflect.DeepEqual(got, want) {
			t.Logf("ClientShares_records: %v, want: %v", got, want)
		}
	}

	// create new client's share
	err = db.InsertClientShare("exp1", "c2", shares)
	if err != nil {
		t.Log(err)
	} else {
		records, err := db.GetClientsSharesPerExperiment("exp1")
		if err != nil {
			t.Log(err)
		} else {
			got, want := len(records), 2
			if got != want {
				t.Logf("num_ClientShares_records=%v, want %v", got, want)
			}
		}
	}

	//update client share
	shares, _ = json.Marshal([][]int{{0, 0, 0}})
	err = db.UpdateClientShare("exp2", "c2", shares)
	if err != nil {
		t.Log(err)
	}

	//DeleteDB("test.db")

}

func TestValidClient(t *testing.T) {
	db := NewDB("test")

	//create a new client's share
	shares, _ := json.Marshal([][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}})
	err := db.InsertClientShare("exp1", "c1", shares)
	if err != nil {
		t.Log(err)
	}

	// create new client's share
	err = db.InsertClientShare("exp1", "c2", shares)
	if err != nil {
		t.Log(err)
	}

	// create new client's share for experiment 2
	err = db.InsertClientShare("exp2", "c2", shares)
	if err != nil {
		t.Log(err)
	}

	//create valid client
	err = db.InsertValidClient("exp1", "c1")
	if err != nil {
		t.Log(err)
	}

	err = db.InsertValidClient("exp2", "c2")
	if err != nil {
		t.Log(err)
	}

	clients, err := db.GetValidClientsPerExperiment("exp1")
	if err != nil {
		t.Log(err)
	} else {
		got, want := len(clients), 1
		if got != want {
			t.Logf("num_valid_clients=%v, want %v", got, want)
		}
	}

	//get valid client share
	records, err := db.GetValidClientShares("exp1")
	if err != nil {
		t.Log(err)
	} else {
		got, want := len(records), 2
		if got != want {
			t.Logf("num_vaid_client_shares=%v, want %v", got, want)
		}
		got1, want1 := records[0].Client_ID, "c1"
		if got1 != want1 {
			t.Logf("client_ID=%v, want %v", got, want)
		}
		got1, want1 = records[1].Client_ID, "c2"
		if got1 != want1 {
			t.Logf("client_ID=%v, want %v", got, want)
		}

	}

	err = db.DeleteValidClient("exp1", "c1")
	if err != nil {
		t.Fatal(err)
	} else {
		clients, err := db.GetValidClientsPerExperiment("exp1")
		if err != nil {
			t.Log(err)
		}
		got, want := len(clients), 0
		if got != want {
			t.Logf("num_valid_clients=%v, want %v", got, want)
		}
	}

	//DeleteDB("test.db")

}

func TestComplaint(t *testing.T) {
	db := NewDB("test")

	//create a new complaint
	err := db.InsertComplaint("exp1", "s1", "c1", true, []byte("0"))
	if err != nil {
		t.Fatal(err)
	}

	//create new complaint
	err = db.InsertComplaint("exp1", "s1", "c2", false, []byte("0"))
	if err != nil {
		t.Fatal(err)
	} else {
		comp, err := db.GetComplaintsPerServer("exp1", "s1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(comp), 2
			if got != want {
				t.Fatalf("num_complaint=%v, want %v", got, want)
			}
		}
	}

	//create new complaint
	err = db.InsertComplaint("exp1", "s2", "c1", false, []byte("0"))
	if err != nil {
		t.Fatal(err)
	} else {
		comp, err := db.GetComplaintsPerClient("exp1", "c1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(comp), 2
			if got != want {
				t.Fatalf("num_complaint=%v, want %v", got, want)
			}
		}
	}

	comp, err := db.GetNoComplain("exp1", "c1")
	if err != nil {
		t.Fatal(err)
	} else {
		got, want := len(comp), 1
		if got != want {
			t.Fatalf("num_no_complain=%v, want %v", got, want)
		}
	}

	//create client
	_ = db.InsertClient("exp1", "c1")
	_ = db.InsertClient("exp1", "c2")
	_ = db.InsertComplaint("exp1", "s1", "c3", true, []byte("0"))
	_ = db.InsertComplaint("exp1", "s1", "c4", true, []byte("0"))
	dropout, err := db.GetDropoutClient("exp1")
	if err != nil {
		t.Fatal(err)
	} else {
		got, want := len(dropout), 2
		if got != want {
			t.Fatalf("num_dropout=%v, want %v", got, want)
		}
		got1, want1 := dropout[0], "c3"
		if got1 != want1 {
			t.Fatalf("num_dropout=%v, want %v", got1, want1)
		}
		got2, want2 := dropout[1], "c4"
		if got2 != want2 {
			t.Fatalf("num_dropout=%v, want %v", got2, want2)
		}

	}

	DeleteDB("test.db")

}
