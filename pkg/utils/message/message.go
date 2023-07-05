package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"example.com/SMC/pkg/packed"
)

type ClientRequest struct {
	Exp_ID       string       `json:"Exp_ID"`
	Client_ID    string       `json:"Client_ID"`
	Token        string       `json:"Token"`
	Secret_Share packed.Share `json:"Secret_Share"`
	Timestamp    string       `json:"Timestamp"`
	//Proof       string       `json:"Proof"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

type ClientRegistry struct {
	Exp_ID    string `json:"Exp_ID"`
	Client_ID string `json:"Client_ID"`
	Token     string `json:"Token"`
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
}

type Writer interface {
	WriteJson() []byte
}

type Reader interface {
	ReadJson(req *http.Request)
}

func (c *ClientRequest) WriteJson() []byte {
	msg := &ClientRequest{
		Exp_ID:       c.Exp_ID,
		Client_ID:    c.Client_ID,
		Secret_Share: c.Secret_Share,
		Timestamp:    c.Timestamp,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall client request: %s", err)
	}

	return message
}

func (c *ClientRegistry) WriteJson() []byte {
	msg := &ClientRequest{
		Exp_ID:    c.Exp_ID,
		Client_ID: c.Client_ID,
		Token:     c.Token,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall client registration: %s", err)
	}

	return message
}

func (op *OutputPartyRequest) WriteJson() []byte {
	msg := &OutputPartyRequest{
		Exp_ID: op.Exp_ID,
		Due:    op.Due,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall output party request: %s", err)
	}

	return message
}

func (s *ServerRequest) WriteJson() []byte {
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

func Send(address string, data []byte) {
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("impossible to build http post request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("impossible to send http request: %s", err)
	}

	log.Printf("response Status:%s", res.Status)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if len(body) > 0 {
		fmt.Println("response Body:", string(body))
	}

}
