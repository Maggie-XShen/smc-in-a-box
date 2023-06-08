package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Client struct {
	Exp_ID      string
	Client_ID   string
	Share_Index int
	Share_Value int
}

type Experiment struct {
	Exp_ID    string `json:"Exp_ID"`
	Due       string `json:"Due"`
	Completed bool
}

type Server struct {
	Server_ID   string
	Share_Index int
	URL         string //output party URL
}

func NewServer(id string, url string, index int) *Server {
	return &Server{Server_ID: id, URL: url, Share_Index: index}
}

func (s *Server) waitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {
		// this gets called every second
		var experiments []Experiment
		r := db.Find(&experiments, "completed=?", false)
		if r.Error != nil {
			panic(r.Error)
		}

		for _, exp := range experiments {

			// TODO: check date

			var clients []Client
			r := db.Find(&clients, "exp_id = ?", exp.Exp_ID)
			if r.Error != nil {
				panic(r.Error)
			}

			// check if all registered clients send their share or time is pased due
			if r.RowsAffected >= 3 { //TODO: change condition to r.RowsAffected >= numOfClients or passDue
				// sum up the shares
				sumSharesValue := s.addShares(clients)

				// send to output party
				msg := utils.Server_Msg{Exp_ID: exp.Exp_ID, Server_ID: s.Server_ID, Sum_Shares: packed.Share{Index: s.Share_Index, Value: sumSharesValue}, Timestamp: time.Now().Format("2006-01-02")}
				fmt.Printf("%+v\n", msg)
				//writer := &msg
				//utils.Send(s.URL, writer.WriteToJson())

				//set exp to completed
				r = db.Model(&exp).Where("exp_ID = ?", exp.Exp_ID).Update("Completed", true)
				if r.Error != nil {
					panic(r.Error)
				}

			}

		}

		// Todo: check if experiment is over and do something (e.g., remove exp and clients information from DB)
	}
}

func (s *Server) clientDataHandler(rw http.ResponseWriter, req *http.Request) {
	client_msg := utils.ReadClientMsg(req)

	var exp Experiment
	var client Client
	checkExp := db.Find(&exp, "exp_id = ?", client_msg.Exp_ID)
	checkClient := db.Find(&client, "exp_id = ? and client_id = ?", client_msg.Exp_ID, client_msg.Client_ID)
	//fmt.Printf("%v\n", checkClient.RowsAffected)

	//Todo: check if client share passes due of experiment and bad events of client share
	if checkExp.RowsAffected == 0 {
		log.Println("Experiment does not exist!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	} else if checkClient.RowsAffected != 0 {
		log.Println("Client record is already exists!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	} else {
		c := Client{
			Exp_ID:      client_msg.Exp_ID,
			Client_ID:   client_msg.Client_ID,
			Share_Index: client_msg.Secret_Share.Index,
			Share_Value: client_msg.Secret_Share.Value,
		}
		db.Create(&c)
	}

	/**
		var users []Client
		result := db.Find(&users)
		fmt.Printf("%+v\n", users)
		fmt.Printf("%v", result.RowsAffected)
	    **/

	// send back result to the client
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

}

func (s *Server) addShares(clients []Client) int {
	sumOfShares := 0

	for _, client := range clients {
		sumOfShares += client.Share_Value
	}
	return sumOfShares
}

func main() {
	//Todo: read parameters from config file and command
	port := flag.String("port", ":8080", "the port on which the server will listen")
	sid := flag.String("sid", "s1", "server ID")
	index := flag.Int("index", 1, "share index")

	flag.Parse()

	exp := &Experiment{ // Todo: set experiments information from output party message
		Exp_ID:    "exp1",
		Due:       "2023-06-01",
		Completed: false,
	}

	// open a database
	var err error
	db_name := fmt.Sprintf("server-%s.db", *sid)
	db, err = gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		panic("failed to connect database") // Todo: change to log
	}
	log.Println("Connection to Database Established")

	db.AutoMigrate(&Experiment{})

	db.Create(&exp)

	db.AutoMigrate(&Client{})

	server := NewServer(*sid, "http://127.0.0.1:8085/serverDataSubmit/", *index)
	http.HandleFunc("/clientDataSubmit/", server.clientDataHandler)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go server.waitForEndOfExperiment(ticker)

	log.Fatal(http.ListenAndServe(*port, nil)) // Todo: read port number from file

}
