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

func TestCreateClient(t *testing.T) {
	db, err := SetupDatabase("test_client")
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	// migrate tables
	storage := repository.NewStorage(db)

	ncs := NewClientService(storage)

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
		}, want: errors.New("client record already exists when server creates client share record")},
		{name: "return an error because experiment does not exist", request: message.ClientRequest{
			Exp_ID:       "exp3",
			Client_ID:    "c1",
			Token:        "tk1",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		}, want: errors.New("experiment does not exist when server creates client share record")},
		{name: "should return error because client share submission passed due", request: message.ClientRequest{
			Exp_ID:       "exp1",
			Client_ID:    "c3",
			Token:        "tk3",
			Secret_Share: packed.Share{Index: 1, Value: 2},
			Timestamp:    time.Now().Add(time.Hour * 20).Format("2006-01-02 15:04:05"),
		}, want: errors.New("client share submission passed due when server creates client share record")},
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
		}, want: errors.New("experiment already exists when server creates experiment record")},
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
