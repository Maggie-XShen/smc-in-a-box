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

type Configuration struct {
	N int
	T int
	K int
	Q int
}

type Experiment struct {
	Exp_ID    string `json:"Exp_ID"`
	Due       string `json:"Due"`
	Completed bool
}

type Server struct {
	Exp_ID     string
	Server_ID  string
	Sum_Shares packed.Share `json:"Sum_Shares"`
}

type OutputParty struct {
	Server_ID string
}

func NewOutputParty(id string) *OutputParty {
	return &OutputParty{Server_ID: id}
}

func (op *OutputParty) reveal(shares []packed.Share, n, t, k, q int) ([]int, error) {
	//Todo: read parameters for packed secret sharing from file
	npss, err := packed.NewPackedSecretSharing(n, t, k, q)

	result, err := npss.Reconstruct(shares)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

	return result, nil
}

func (op *OutputParty) waitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {
		// this gets called every second
		var experiments []Experiment
		r := db.Find(&experiments, "completed=?", false)
		if r.Error != nil {
			panic(r.Error)
		}

		for _, exp := range experiments {

			var servers []Server
			r := db.Find(&servers, "exp_id = ?", exp.Exp_ID)
			if r.Error != nil {
				panic(r.Error)
			}

			// check if all servers send their share
			if r.RowsAffected == 3 { // TODO: get number of servers from config gile
				// reconstruct sum of secrets
				shares := make([]packed.Share, 3)
				for _, server := range servers {
					shares = append(shares, server.Sum_Shares)
				}

				op.reveal(shares, n, t, k, q)

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

func (op *OutputParty) serverDataHandler(rw http.ResponseWriter, req *http.Request) {
	server_msg := utils.ReadServerMsg(req)

	//Todo: check validity of server

	//store server data to database
	var exp Experiment
	var server Server
	checkExp := db.Find(&exp, "exp_id = ?", server_msg.Exp_ID)
	checkServer := db.Find(&server, "exp_id = ? and server_id = ?", server_msg.Exp_ID, server_msg.Server_ID)

	if checkExp.RowsAffected == 0 {
		log.Println("Experiment does not exist!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	} else if checkServer.RowsAffected != 0 {
		log.Println("Server record is already exists!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	} else {
		s := Server{
			Exp_ID:     server_msg.Exp_ID,
			Server_ID:  server_msg.Server_ID,
			Sum_Shares: server_msg.Sum_Shares,
		}
		db.Create(&s)
	}

	// send back result to the client
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

}

func main() {
	port := flag.String("port", ":8080", "the port on which the server will listen")
	oid := flag.String("oid", "o1", "output party ID")

	// Todo: set experiments information from file
	exp := &Experiment{ // Todo: set experiments information from output party message
		Exp_ID:    "exp1",
		Due:       "2023-06-01",
		Completed: false,
	}

	// open a database
	var err error
	db_name := fmt.Sprintf("OutputParty-%s.db", *oid)
	db, err = gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database") // Todo: change to log
	}
	log.Println("Connection to Database Established")

	db.AutoMigrate(&Experiment{})

	db.Create(&exp)

	db.AutoMigrate(&Server{})

	outputParty := NewOutputParty(*oid)
	http.HandleFunc("/serverDataSubmit/", outputParty.serverDataHandler)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go outputParty.waitForEndOfExperiment(ticker)

	// Todo: read port number from file
	log.Fatal(http.ListenAndServe(*port, nil))

}
