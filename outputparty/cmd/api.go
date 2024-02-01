package main

import (
	"errors"

	"example.com/SMC/outputparty/public/utils"
	"example.com/SMC/outputparty/sqlstore"
)

type ServerService struct {
	store *sqlstore.SqlStore
}

type ExperimentService struct {
	store *sqlstore.SqlStore
}

func NewServerService(s *sqlstore.SqlStore) *ServerService {
	return &ServerService{store: s}
}

func NewExperimentService(s *sqlstore.SqlStore) *ExperimentService {
	return &ExperimentService{store: s}
}

func (ss *ServerService) CreateServerShare(request utils.ServerRequest) error {
	//Todo: check validity of server request

	exp, err := ss.store.GetExperiment(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when output party creates server's shares record")
	}

	//insert share to server share table
	for _, sh := range request.Shares {
		err = ss.store.InsertServerShare(request.Exp_ID, request.Server_ID, sh.Index, sh.Value)
		if err != nil {
			return err
		}
	}

	return nil

}

func (e *ExperimentService) CreateExperiment(request utils.OutputPartyRequest) error {
	err := e.store.InsertExperiment(request.Exp_ID, request.Due, request.Owner)
	if err != nil {
		return err
	}

	return nil
}
