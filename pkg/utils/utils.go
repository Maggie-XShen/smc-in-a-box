package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"example.com/SMC/pkg/packed"
)

type Client_Msg struct {
	Exp_ID       string       `json:"Exp_ID"`
	Client_ID    string       `json:"Client_ID"`
	Secret_Share packed.Share `json:"Secret_Share"`
	Timestamp    string       `json:"Timestamp"`
	//Proof       string       `json:"Proof"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

type Server_Msg struct {
	Exp_ID     string       `json:"Exp_ID "`
	Server_ID  string       `json:"Server_ID"`
	Sum_Shares packed.Share `json:"Sum_Shares"`
	Timestamp  string       `json:"Timestamp"`
}

type Writer interface {
	WriteToJson() []byte
}

func (c *Client_Msg) WriteToJson() []byte {
	msg := &Client_Msg{
		Exp_ID:       c.Exp_ID,
		Client_ID:    c.Client_ID,
		Secret_Share: c.Secret_Share,
		Timestamp:    c.Timestamp, // Todo: decide time format
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("impossible to marshall response: %s", err)
	}

	return message
}

func (s *Server_Msg) WriteToJson() []byte {
	msg := &Server_Msg{
		Exp_ID:     s.Exp_ID,
		Server_ID:  s.Server_ID,
		Sum_Shares: s.Sum_Shares,
		Timestamp:  s.Timestamp, // Todo: decide time format
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("impossible to marshall response: %s", err)
	}

	return message
}

func ReadClientMsg(req *http.Request) *Client_Msg {
	decoder := json.NewDecoder(req.Body)
	var t Client_Msg
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	return &t
}

func ReadServerMsg(req *http.Request) *Server_Msg {
	decoder := json.NewDecoder(req.Body)
	var t Server_Msg
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	return &t
}

func Send(address string, data []byte) {
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("impossible to build request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("impossible to send request: %s", err)
	}
	log.Printf("response Status:%s", res.Status)

	//defer res.Body.Close()
	//body, _ := io.ReadAll(res.Body)
	//fmt.Println("response Body:", string(body))

}
