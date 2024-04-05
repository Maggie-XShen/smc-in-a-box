package main

import (
	"errors"

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

	//insert share to server share table
	for input_index, party := range request.Shares {
		for _, share := range party.Shares {
			err = ss.store.InsertServerShare(request.Exp_ID, request.Server_ID, input_index, share.Index, share.Value)
			if err != nil {
				return err
			}
		}

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
