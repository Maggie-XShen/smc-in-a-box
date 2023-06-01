package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/SMC/pkg/packed"
)

type Experiment struct {
	EID          string                  `json:"EID"`
	Due          string                  `json:"Due"`
	Participants map[string]packed.Share `json:"Participants"` //key is CID
}

type Server struct {
	SID  string
	Exps map[string]Experiment
}

type Msg struct {
	EID       string       `json:"EID"`
	SID       string       `json:"SID"`
	SumShares packed.Share `json:"SumShares"`
	Timestamp string       `json:"Timestamp"`
}

func NewServer(id string) *Server {
	return &Server{SID: id, Exps: make(map[string]Experiment)}
}

func (s *Server) clientDataHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	t := struct {
		EID         string       `json:"EID"`
		CID         string       `json:"CID"`
		SecretShare packed.Share `json:"SecretShare"`
		Timestamp   string       `json:"Timestamp"`
	}{}
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Println(t)

	// Todo: check bad events of client share and client share does not pass due of experiment

	//store client data to database
	exp, exists := s.Exps[t.EID]
	if exists {
		_, exists := exp.Participants[t.CID]
		if !exists {
			exp.Participants[t.CID] = t.SecretShare
		} else {
			log.Println("Client share is discarded because it already exists!")
		}
	} else {
		log.Println("Experiment does not exist!")
	}
	fmt.Println(s.Exps)
}

func (s *Server) addShares(eid string, serverIndex int) (packed.Share, error) {
	exp := s.Exps[eid]

	var sumShares packed.Share
	sumShares.Index = serverIndex
	sumShares.Value = 0
	for _, share := range exp.Participants {
		sumShares.Value += share.Value
	}
	return sumShares, nil
}

func (s *Server) write(eid string, sid string, sumShares packed.Share) ([]byte, error) {
	currentTime := time.Now()
	msg := &Msg{
		EID:       eid,
		SID:       sid,
		SumShares: sumShares,
		Timestamp: currentTime.Format("2006-01-02"), // Todo: decide time format
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("impossible to marshall response: %s", err)
	}

	return message, nil
}

/**
func (s *Server) send(address string, data []byte) {
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

}**/

func main() {

	port := flag.String("port", ":8080", "the port on which the server will listen")
	sid := flag.String("sid", "s1", "server ID")
	//index := flag.Int("index", 10, "aggregated share's index")

	flag.Parse()

	server := NewServer(*sid)

	// Todo: set experiments information from file
	server.Exps["exp1"] = Experiment{EID: "exp1", Due: "2023-06-01", Participants: make(map[string]packed.Share)}

	http.HandleFunc("/clientDataSubmit/", server.clientDataHandler)

	// Todo: read port number from file
	log.Fatal(http.ListenAndServe(*port, nil))

	// Todo: decide when server aggregates shares and send message to output party after due
	//go server.addShares("exp1", *index)
	//server.write()
	//server.send()
}
