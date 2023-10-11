package main

import (
	"errors"

	"example.com/SMC/outputparty/public/utils"
	"example.com/SMC/outputparty/sqlstore"
)

type ServerService struct {
	store *sqlstore.SqlStore
}

func NewServerService(s *sqlstore.SqlStore) *ServerService {
	return &ServerService{store: s}
}

func (ss *ServerService) CreateServer(request utils.ServerRequest) error {
	//Todo: check validity of server

	exp, err := ss.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when output party creates server sumshare record")
	}

	server, err := ss.store.GetServerComputation(request.Exp_ID, request.Server_ID)
	if err != nil {
		return err
	}
	if *server != (sqlstore.ServerComputation{}) {
		return errors.New("server record already exists when output party creates server sumshare record")
	}
	err = ss.store.InsertServerComputation(request)
	if err != nil {
		return err
	}

	return nil

}

type ExperimentService struct {
	store *sqlstore.SqlStore
}

func NewExperimentService(s *sqlstore.SqlStore) *ExperimentService {
	return &ExperimentService{store: s}
}

func (e *ExperimentService) CreateExp(request utils.OutputPartyRequest) error {
	exp, err := e.store.GetExp(request.Exp_ID)

	if err != nil {
		return err
	}

	if *exp != (sqlstore.Experiment{}) {
		return errors.New("experiment already exists when output party creates experiment record")
	}

	err = e.store.InsertExp(request)
	if err != nil {
		return err
	}

	return nil
}
