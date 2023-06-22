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
	"example.com/SMC/pkg/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateExperiment(t *testing.T) {
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

	nes := api.NewExperimentService(*storage)

	due := time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")

	tests := []struct {
		name    string
		request message.OutputPartyRequest
		want    error
	}{
		{name: "create exp1 record successfully", request: message.OutputPartyRequest{
			Exp_ID: "exp1",
			Due:    due,
		}, want: nil},
		{name: "create exp2 record successfully", request: message.OutputPartyRequest{
			Exp_ID: "exp2",
			Due:    due,
		}, want: nil},
		{name: "return an error because experiment already exists ", request: message.OutputPartyRequest{
			Exp_ID: "exp1",
			Due:    due,
		}, want: errors.New("experiment already exists")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := nes.CreateExp(test.request)
			//fmt.Printf("%+v\n", test.request)
			if !reflect.DeepEqual(err, test.want) {
				t.Errorf("test: %v failed. got: %v, wanted: %v", test.name, err, test.want)
			}
		})
	}

}
