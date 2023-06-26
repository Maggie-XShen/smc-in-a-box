package main

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/repository"
	"example.com/SMC/pkg/utils/message"
)

func TestCreateServer(t *testing.T) {
	db, err := SetupDatabase("test_server")
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	// migrate tables
	storage := repository.NewStorage(db)

	nss := NewServerService(storage)

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
		}, want: errors.New("server record already exists when output party creates server sumshare record")},
		{name: "return an error because experiment does not exist", request: message.ServerRequest{
			Exp_ID:     "exp3",
			Server_ID:  "s1",
			Sum_Shares: packed.Share{Index: 1, Value: 2},
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		}, want: errors.New("experiment does not exist when output party creates server sumshare record")},
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

func TestCreateExperiment(t *testing.T) {
	db, err := SetupDatabase("test_exp")
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	// migrate tables
	storage := repository.NewStorage(db)

	nes := NewExperimentService(storage)

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
		}, want: errors.New("experiment already exists when output party creates experiment record")},
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
