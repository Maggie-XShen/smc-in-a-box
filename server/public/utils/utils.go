package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/pkg/packed"
	"gorm.io/datatypes"
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

type ClientSet struct {
	Exp_ID    string
	Server_ID string
	Clients   datatypes.JSON
}

type ServerRequest struct {
	Exp_ID     string       `json:"Exp_ID "`
	Server_ID  string       `json:"Server_ID"`
	Sum_Shares packed.Share `json:"Sum_Shares"`
	Timestamp  string       `json:"Timestamp"`
}

type OutputPartyRequest struct {
	Exp_ID string `json:"Exp_ID"`
	Due    string `json:"Due"`
	Owner  string `json:"Owner"`
}

type Reader interface {
	ReadJson(req *http.Request)
}

func GenerateClientSetRecord(exp_id string, server_id string, clients []string) ClientSet {
	clientsJSON, err := json.Marshal(clients)
	if err != nil {
		log.Fatalf("failed to marshal client set to JSON")
	}
	request := ClientSet{
		Exp_ID:    exp_id,
		Server_ID: server_id,
		Clients:   clientsJSON,
	}
	return request
}

func (c *ClientSet) ToJson() []byte {
	msg := &ClientSet{
		Exp_ID:    c.Exp_ID,
		Server_ID: c.Server_ID,
		Clients:   c.Clients,
	}

	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall client space: %s", err)
	}

	return message
}

func (s *ServerRequest) ToJson() []byte {
	msg := &ServerRequest{
		Exp_ID:     s.Exp_ID,
		Server_ID:  s.Server_ID,
		Sum_Shares: s.Sum_Shares,
		Timestamp:  s.Timestamp,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall server request: %s", err)
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

	/**
	fmt.Printf("Size of Total: %d bytes\n", unsafe.Sizeof(t))
	fmt.Printf("Size of Proof: %d bytes\n", unsafe.Sizeof(t.Proof))
	fmt.Printf("Size of Proof: %d bytes\n", unsafe.Sizeof(t.Proof.MerkleRoot))
	fmt.Printf("Size of q_code: %d bytes\n", unsafe.Sizeof(t.Proof.CodeTest))
	fmt.Printf("Size of q_quadra: %d bytes\n", unsafe.Sizeof(t.Proof.QuadraTest))
	fmt.Printf("Size of q_linear: %d bytes\n", unsafe.Sizeof(t.Proof.LinearTest))
	fmt.Printf("Size of open columns: %d bytes\n", unsafe.Sizeof(t.Proof.ColumnTest))
	**/

	return t
}

func (c *ClientSet) ReadJson(req *http.Request) ClientSet {
	decoder := json.NewDecoder(req.Body)
	var t ClientSet
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode client space: %s", err)
	}
	return t
}

func (s *ServerRequest) ReadJson(req *http.Request) ServerRequest {
	decoder := json.NewDecoder(req.Body)
	var t ServerRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode server request: %s", err)
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
