package api_test

import (
	"errors"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"example.com/SMC/pkg/api"
	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateServer(t *testing.T) {
	dbName := "test.db"

	// remove old database
	os.Remove(dbName)

	// open and create a new database
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// migrate tables
	storage := repository.NewStorage(db)
	storage.Migrate()

	nss := api.NewServerService(*storage)

	//add experiment infor
	due := time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")
	storage.CreateExp(message.OutputPartyRequest{Exp_ID: "exp1", Due: due})
	storage.CreateExp(message.OutputPartyRequest{Exp_ID: "exp2", Due: due})

	tests := []struct {
		name    string
		request message.ServerRequest
		want    error
	}{
		{name: "create s1 record of exp1 successfully", request: message.ServerRequest{
			Exp_ID:     "exp1",
			Server_ID:  "s1",
			Sum_Shares: packed.Share{Index: 1, Value: 2},
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		}, want: nil},
		{name: "create s2 record of exp1 successfully", request: message.ServerRequest{
			Exp_ID:     "exp1",
			Server_ID:  "s2",
			Sum_Shares: packed.Share{Index: 1, Value: 2},
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		}, want: nil},
		{name: "create s1 record of exp2 successfully", request: message.ServerRequest{
			Exp_ID:     "exp2",
			Server_ID:  "s1",
			Sum_Shares: packed.Share{Index: 1, Value: 2},
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		}, want: nil},
		{name: "return an error because server already exists ", request: message.ServerRequest{
			Exp_ID:     "exp1",
			Server_ID:  "s1",
			Sum_Shares: packed.Share{Index: 1, Value: 2},
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		}, want: errors.New("server record already exists")},
		{name: "return an error because experiment does not exist", request: message.ServerRequest{
			Exp_ID:     "exp3",
			Server_ID:  "s1",
			Sum_Shares: packed.Share{Index: 1, Value: 2},
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		}, want: errors.New("experiment does not exist")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := nss.CreateServer(test.request)
			//fmt.Printf("%+v\n", test.request)
			if !reflect.DeepEqual(err, test.want) {
				t.Errorf("test: %v failed. got: %v, wanted: %v", test.name, err, test.want)
			}
		})
	}
}
