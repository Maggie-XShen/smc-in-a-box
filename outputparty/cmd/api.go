package main

import (
	"encoding/json"
	"errors"
	"log"

	"example.com/SMC/outputparty/sqlstore"
)

type ServerService struct {
	store *sqlstore.DB
}

type ExperimentService struct {
	store *sqlstore.DB
}

func NewServerService(s *sqlstore.DB) *ServerService {
	return &ServerService{store: s}
}

func NewExperimentService(s *sqlstore.DB) *ExperimentService {
	return &ExperimentService{store: s}
}

func (ss *ServerService) CreateServerShare(request AggregatedShareRequest) error {
	//Todo: check validity of server request
	exp, err := ss.store.GetExperiment(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when output party creates server's shares record")
	}

	log.Printf("outputparty received server shares from %s\n", request.Server_ID)

	shares, err := json.Marshal(request.Shares)
	if err != nil {
		log.Fatal(err)
	}

	//insert share to server share table
	err = ss.store.InsertServerShare(request.Exp_ID, request.Server_ID, shares)
	if err != nil {
		return err
	}

	return nil

}

func (e *ExperimentService) CreateExperiment(exp Experiment) error {
	err := e.store.InsertExperiment(exp.Exp_ID, exp.ClientShareDue, exp.ServerShareDue)
	if err != nil {
		return err
	}

	return nil
}
