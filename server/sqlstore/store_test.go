package sqlstore

import (
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
			t.Fatal(err)
		} else {
			got, want := len(clients), 1
			if got != want {
				t.Fatalf("num_clients=%v, want %v", got, want)
			}
		}
	}

	// create same client for experiment 1
	err = db.InsertClient("exp1", "c1")
	if err != nil {
		t.Fatal(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(clients), 1
			if got != want {
				t.Fatalf("num_clients=%v, want %v", got, want)
			}
		}
	}

	// create second client for experiment 1
	err = db.InsertClient("exp1", "c2")
	if err != nil {
		t.Fatal(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(clients), 2
			if got != want {
				t.Fatalf("num_clients=%v, want %v", got, want)
			}
		}
	}

	// create new client for experiment 2
	err = db.InsertClient("exp2", "c1")
	if err != nil {
		t.Fatal(err)
	} else {
		clients, err := db.GetClientsPerExperiment("exp2")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(clients), 1
			if got != want {
				t.Fatalf("num_clients=%v, want %v", got, want)
			}
		}
	}

	DeleteDB("test.db")

}

func TestInsertClientShare(t *testing.T) {
	db := NewDB("test")

	//create a new client's share
	err := db.InsertClientShare("exp1", "c1", 1, 1, 12)
	if err != nil {
		t.Fatal(err)
	}

	// create client share which already exist
	err = db.InsertClientShare("exp1", "c1", 1, 1, 15)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetClientShares("exp1", "c1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_client_shares=%v, want %v", got, want)
			}
			got, want = shares[0].Value, 12
			if got != want {
				t.Fatalf("client_share_value=%v, want %v", got, want)
			}

		}
	}

	// create second client share
	err = db.InsertClientShare("exp1", "c1", 1, 2, 20)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetClientShares("exp1", "c1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 2
			if got != want {
				t.Fatalf("num_client_shares=%v, want %v", got, want)
			}
		}
	}

	// create new client's share
	err = db.InsertClientShare("exp1", "c2", 1, 1, 23)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetClientsSharesPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 3
			if got != want {
				t.Fatalf("num_shares=%v, want %v", got, want)
			}
		}
	}

	// create new client's share for experiment 2
	err = db.InsertClientShare("exp2", "c2", 1, 1, 23)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetClientsSharesPerExperiment("exp2")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_shares=%v, want %v", got, want)
			}
		}
	}

	//update client share
	err = db.UpdateClientShare("exp2", "c2", 1, 1, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		shares, err := db.GetClientShares("exp2", "c2")
		if err != nil {
			t.Fatal(err)
		} else {
			got, want := len(shares), 1
			if got != want {
				t.Fatalf("num_client_shares=%v, want %v", got, want)
			}
			got, want = shares[0].Value, 100
			if got != want {
				t.Fatalf("client_share_value=%v, want %v", got, want)
			}
		}
	}

	DeleteDB("test.db")

}

func TestValidClient(t *testing.T) {
	db := NewDB("test")

	//create a new client's share
	err := db.InsertClientShare("exp1", "c1", 1, 1, 12)
	if err != nil {
		t.Fatal(err)
	}

	// create second client share
	err = db.InsertClientShare("exp1", "c1", 1, 2, 20)
	if err != nil {
		t.Fatal(err)
	}

	// create new client's share
	err = db.InsertClientShare("exp1", "c2", 1, 1, 23)
	if err != nil {
		t.Fatal(err)
	}

	// create new client's share for experiment 2
	err = db.InsertClientShare("exp2", "c2", 1, 1, 23)
	if err != nil {
		t.Fatal(err)
	}

	//create valid client
	err = db.InsertValidClient("exp1", "c1")
	if err != nil {
		t.Fatal(err)
	}

	err = db.InsertValidClient("exp2", "c2")
	if err != nil {
		t.Fatal(err)
	}

	clients, err := db.GetValidClientsPerExperiment("exp1")
	if err != nil {
		t.Fatal(err)
	} else {
		got, want := len(clients), 1
		if got != want {
			t.Fatalf("num_valid_clients=%v, want %v", got, want)
		}
	}

	//get valid client share
	shares, err := db.GetValidClientShares("exp1")
	if err != nil {
		t.Fatal(err)
	} else {
		got, want := len(shares), 2
		if got != want {
			t.Fatalf("num_vaid_client_shares=%v, want %v", got, want)
		}
		got1, want1 := shares[0].Client_ID, "c1"
		if got1 != want1 {
			t.Fatalf("client_ID=%v, want %v", got, want)
		}
		got1, want1 = shares[1].Client_ID, "c1"
		if got1 != want1 {
			t.Fatalf("client_ID=%v, want %v", got, want)
		}

	}

	err = db.DeleteValidClient("exp1", "c1")
	if err != nil {
		t.Fatal(err)
	} else {
		clients, err := db.GetValidClientsPerExperiment("exp1")
		if err != nil {
			t.Fatal(err)
		}
		got, want := len(clients), 0
		if got != want {
			t.Fatalf("num_valid_clients=%v, want %v", got, want)
		}
	}

	DeleteDB("test.db")

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
