package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

type DolevComplaintRequest struct {
	Round_ID   int              `json:"Round_ID"`
	Server_ID  string           `json:"Server_ID"`
	Msg        ComplaintRequest `json:"Msg"`
	Signatures []Signature      `json:"Signatures"`
}

type Signature struct {
	Sig       []byte `json:"Sig"`
	Server_ID string `json:"Server_ID"`
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

type DolevMaskedShareRequest struct {
	Round_ID   int                `json:"Round_ID"`
	Server_ID  string             `json:"Server_ID"`
	Msg        MaskedShareRequest `json:"Msg"`
	Signatures []Signature        `json:"Signatures"`
}

type AggregatedShareRequest struct {
	Exp_ID    string      `json:"Exp_ID "`
	Server_ID string      `json:"Server_ID"`
	Shares    []rss.Share `json:"Shares"`
	Timestamp string      `json:"Timestamp"`
}

type Experiment struct {
	Exp_ID            string `json:"Exp_ID"`
	ClientShareDue    string `json:"ClientShareDue"`
	ComplaintDue      string `json:"ComplaintDue"`
	ShareBroadcastDue string `json:"ShareBroadcastDue"`
	Owner             string `json:"Owner"`
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

func (dcr *DolevComplaintRequest) ToJson() []byte {
	msg := &DolevComplaintRequest{
		Round_ID:   dcr.Round_ID,
		Server_ID:  dcr.Server_ID,
		Msg:        dcr.Msg,
		Signatures: dcr.Signatures,
	}

	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall delov complaints request: %s", err)
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

func (dr *DolevMaskedShareRequest) ToJson() []byte {
	msg := &DolevMaskedShareRequest{
		Round_ID:   dr.Round_ID,
		Server_ID:  dr.Server_ID,
		Msg:        dr.Msg,
		Signatures: dr.Signatures,
	}

	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall delov complaints request: %s", err)
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

func (dc *DolevComplaintRequest) ReadJson(req *http.Request) DolevComplaintRequest {
	decoder := json.NewDecoder(req.Body)
	var t DolevComplaintRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode server delov complaint: %s", err)
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

func (dm *DolevMaskedShareRequest) ReadJson(req *http.Request) DolevMaskedShareRequest {
	decoder := json.NewDecoder(req.Body)
	var t DolevMaskedShareRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode server delov masked share: %s", err)
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

func FindMajority(list []int, t int) (int, error) {
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
	if maxCount >= t+1 {
		return list[index], nil
	}

	return 0, fmt.Errorf("reconstruct failed: no majority element")
}

func ReadServerInput(path string) []Experiment {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}

	var items []Experiment
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	return items

}
