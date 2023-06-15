package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/SMC/cmd/outputparty/config"
	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/packed"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var URLs []string

type Experiment struct {
	Exp_ID    string `json:"Exp_ID"`
	Due       string `json:"Due"`
	Completed bool
}

type Server struct {
	Exp_ID         string
	Server_ID      string
	SumShare_Value int
	SumShare_Index int
}

type OutputParty struct {
	OutputParty_ID string
}

func NewOutputParty(id string) *OutputParty {
	return &OutputParty{OutputParty_ID: id}
}

func (op *OutputParty) reveal(shares []packed.Share) ([]int, error) {
	//read parameters for packed secret sharing from file
	conf := config.LoadConfig("config/config.json")
	npss, err := packed.NewPackedSecretSharing(conf.N, conf.T, conf.K, conf.Q)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("test")
	result, err := npss.Reconstruct(shares)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", result)

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
				var shares []packed.Share
				for _, server := range servers {
					shares = append(shares, packed.Share{Value: server.SumShare_Value, Index: server.SumShare_Index})
				}

				fmt.Printf("%+v\n", shares)

				op.reveal(shares)

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
	server_msg := message.ReadServerMsg(req)
	fmt.Printf("%v\n", server_msg)

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
			Exp_ID:         server_msg.Exp_ID,
			Server_ID:      server_msg.Server_ID,
			SumShare_Value: server_msg.Sum_Shares.Value,
			SumShare_Index: server_msg.Sum_Shares.Index,
		}
		db.Create(&s)
	}

	// send back result to the server
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

}

func (op *OutputParty) ConnectDB(oid string) {
	// open a database
	var err error
	db_name := fmt.Sprintf("OutputParty-%s.db", oid)
	db, err = gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		panic("failed to connect database") // Todo: change to log
	}
	log.Println("Connection to Database Established")

	db.AutoMigrate(&Experiment{})

	db.AutoMigrate(&Server{})
}

func (op *OutputParty) HandelExpInfor() {

	var exp Experiment
	checkExp := db.Find(&exp, "exp_id = ?", exp.Exp_ID)

	if checkExp.RowsAffected != 0 {
		log.Println("Experiment already exists!")
		return
	}

	// Todo: set experiments information from file
	exp.Exp_ID = "exp1"
	exp.Due = "2023-06-01"
	exp.Completed = false
	db.Create(&exp)

	msg := message.OutputParty_Msg{Exp_ID: exp.Exp_ID, Due: exp.Due, Completed: exp.Completed}
	fmt.Printf("%+v\n", msg)
	writer := &msg
	for _, url := range URLs {
		message.Send(url, writer.WriteToJson())
	}

}

func main() {
	//read configuration
	confpath := flag.String("confpath", "config/config.json", "config file path") // confpath := "config.json"
	exppath := flag.String("exppath", "exp.json", "experiment file path")
	flag.Parse()

	conf := config.LoadConfig(*confpath)
	URLs = conf.URLs

	outputParty := NewOutputParty(conf.OutputParty_ID)
	outputParty.ConnectDB(conf.OutputParty_ID)

	if *exppath != "" {
		outputParty.HandelExpInfor()
	}

	http.HandleFunc("/serverDataSubmit/", outputParty.serverDataHandler)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go outputParty.waitForEndOfExperiment(ticker)

	// Todo: read port number from file
	log.Fatal(http.ListenAndServe(":"+conf.Port, nil))

}
