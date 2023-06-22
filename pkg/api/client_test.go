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

func SetupDatabase(db_name string) (*gorm.DB, error) {
	// open a database
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("Connection to Database Established")

	return db, nil
}

func TestCreateClient(t *testing.T) {
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

	ncs := api.NewClientService(*storage)

	//add experiment infor
	due := time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")
	storage.CreateExp(message.OutputPartyRequest{Exp_ID: "exp1", Due: due})
	storage.CreateExp(message.OutputPartyRequest{Exp_ID: "exp2", Due: due})

	tests := []struct {
		name    string
		request message.ClientRequest
		want    error
	}{
		{name: "create c1 record of exp1 successfully", request: message.ClientRequest{
			Exp_ID:       "exp1",
			Client_ID:    "c1",
			Token:        "tk1",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		}, want: nil},
		{name: "create c2 record of exp1 successfully", request: message.ClientRequest{
			Exp_ID:       "exp1",
			Client_ID:    "c2",
			Token:        "tk2",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		}, want: nil},
		{name: "create c1 record of exp2 successfully", request: message.ClientRequest{
			Exp_ID:       "exp2",
			Client_ID:    "c1",
			Token:        "tk1",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		}, want: nil},
		{name: "return an error because client already exists ", request: message.ClientRequest{
			Exp_ID:       "exp1",
			Client_ID:    "c1",
			Token:        "tk1",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		}, want: errors.New("client record already exists")},
		{name: "return an error because experiment does not exist", request: message.ClientRequest{
			Exp_ID:       "exp3",
			Client_ID:    "c1",
			Token:        "tk1",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		}, want: errors.New("experiment does not exist")},
		{name: "should return error because client share submission passed due", request: message.ClientRequest{
			Exp_ID:       "exp1",
			Client_ID:    "c3",
			Token:        "tk3",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Add(time.Hour * 20).Format("2006-01-02 15:04:05"),
		}, want: errors.New("client share submission passed due")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ncs.CreateClient(test.request)
			//fmt.Printf("%+v\n", test.request)
			if !reflect.DeepEqual(err, test.want) {
				t.Errorf("test: %v failed. got: %v, wanted: %v", test.name, err, test.want)
			}
		})
	}
}
