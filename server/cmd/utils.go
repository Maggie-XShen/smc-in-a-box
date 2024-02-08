package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/pkg/rss"
)

type ClientRequest struct {
	Exp_ID    string       `json:"Exp_ID"`
	Client_ID string       `json:"Client_ID"`
	Token     string       `json:"Token"`
	Proof     ligero.Proof `json:"Proof"`
	Timestamp string       `json:"Timestamp"`
	//Proof       string       `json:"Proof"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

type ClientRegistry struct {
	Exp_ID    string `json:"Exp_ID"`
	Client_ID string `json:"Client_ID"`
	Token     string `json:"Token"`
}

type ComplaintRequest struct {
	Exp_ID     string      `json:"Exp_ID"`
	Server_ID  string      `json:"Server_ID"`
	Complaints []Complaint `json:"Complaints"`
}

type Complaint struct {
	Client_ID string `json:"Client_ID"`
	Complain  bool   `json:"Complain"`
	Root      []byte `json:"Root"`
}

type MaskedShareRequest struct {
	Exp_ID       string        `json:"Exp_ID"`
	Server_ID    string        `json:"Server_ID"`
	MaskedShares []MaskedShare `json:"MaskedShares"`
}

type MaskedShare struct {
	Client_ID   string `json:"Client_ID"`
	Input_Index int    `json:"Input_Index"`
	Index       int    `json:"Index"`
	Value       int    `json:"Value"`
}

type AggregatedShareRequest struct {
	Exp_ID    string      `json:"Exp_ID "`
	Server_ID string      `json:"Server_ID"`
	Shares    []rss.Share `json:"Shares"`
	Timestamp string      `json:"Timestamp"`
}

type OutputPartyRequest struct {
	Exp_ID         string `json:"Exp_ID"`
	ClientShareDue string `json:"ClientShareDue"`
	Owner          string `json:"Owner"`
}

type Reader interface {
	ReadJson(req *http.Request)
}

func (cr *ComplaintRequest) ToJson() []byte {
	msg := &ComplaintRequest{
		Exp_ID:     cr.Exp_ID,
		Server_ID:  cr.Server_ID,
		Complaints: cr.Complaints,
	}

	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall complaints request: %s", err)
	}

	return message
}

func (r *MaskedShareRequest) ToJson() []byte {
	msg := &MaskedShareRequest{
		Exp_ID:       r.Exp_ID,
		Server_ID:    r.Server_ID,
		MaskedShares: r.MaskedShares,
	}

	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall masked share request: %s", err)
	}

	return message
}

func (s *AggregatedShareRequest) ToJson() []byte {
	msg := &AggregatedShareRequest{
		Exp_ID:    s.Exp_ID,
		Server_ID: s.Server_ID,
		Shares:    s.Shares,
		Timestamp: s.Timestamp,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall aggregated share request: %s", err)
	}

	return message
}

func (c *ClientRequest) ReadJson(req *http.Request) ClientRequest {
	decoder := json.NewDecoder(req.Body)
	var t ClientRequest

	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode client request: %s", err)
	}

	return t
}

func (c *ComplaintRequest) ReadJson(req *http.Request) ComplaintRequest {
	decoder := json.NewDecoder(req.Body)
	var t ComplaintRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode server complaint: %s", err)
	}
	return t
}

func (m *MaskedShareRequest) ReadJson(req *http.Request) MaskedShareRequest {
	decoder := json.NewDecoder(req.Body)
	var t MaskedShareRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode masked shares request: %s", err)
	}
	return t
}

func (s *AggregatedShareRequest) ReadJson(req *http.Request) AggregatedShareRequest {
	decoder := json.NewDecoder(req.Body)
	var t AggregatedShareRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode aggregated share request: %s", err)
	}
	return t
}

func (op *OutputPartyRequest) ReadJson(req *http.Request) OutputPartyRequest {
	decoder := json.NewDecoder(req.Body)
	var t OutputPartyRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode experiment request: %s", err)
	}
	return t
}

func FindMajority(list []int) (int, error) {
	maxCount := 0
	index := -1
	n := len(list)
	for i := 0; i < n; i++ {
		count := 0
		for j := 0; j < n; j++ {
			if list[i] == list[j] {
				count++
			}

		}

		// update maxCount if count of
		// current element is greater
		if count > maxCount {
			maxCount = count
			index = i
		}
	}

	// if maxCount is greater than n/2
	// return the corresponding element
	if maxCount > n/2 {
		return list[index], nil
	}

	return 0, fmt.Errorf("reconstruct failed: no majority element")
}
