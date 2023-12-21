package main

import (
	"errors"
	"time"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/server/config"
	"example.com/SMC/server/public/utils"
	"example.com/SMC/server/sqlstore"
)

type ClientService struct {
	store *sqlstore.SqlStore
}

type ServerService struct {
	store *sqlstore.SqlStore
}

type ExperimentService struct {
	store *sqlstore.SqlStore
}

func NewClientService(s *sqlstore.SqlStore) *ClientService {
	return &ClientService{store: s}
}

func NewServerService(s *sqlstore.SqlStore) *ServerService {
	return &ServerService{store: s}
}

func NewExperimentService(s *sqlstore.SqlStore) *ExperimentService {
	return &ExperimentService{store: s}
}

func (c *ClientService) CreateClient(request utils.ClientRequest, cfg *config.Server) error {
	// TODO : do some basic validations, e.g. missing exp_id, client_id, etc.
	exp, err := c.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}

	// TODO : change this check to check if client in the registry table
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when server creates client share record")
	}

	timestamp, _ := time.Parse("2006-01-02 15:04:05", request.Timestamp)
	due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)
	if timestamp.After(due) {
		return errors.New("client share submission passed due when server creates client share record")
	}

	client, err := c.store.GetClient(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}
	if *client != (sqlstore.ClientShare{}) {
		return errors.New("client record already exists when server creates client share record")
	}

	zk, err := ligero.NewLigeroZK(cfg.N_claims, cfg.M, cfg.N, cfg.T, cfg.Q, cfg.N_open)
	if err != nil {
		return err
	}

	//check client's proof
	verify, err := zk.Verify(request.Proof)
	if err != nil {
		return err
	}
	if !verify {
		return errors.New("failed verification of proof!")
	}

	err = c.store.InsertClientShare(request)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClientService) CreateClientRegistry(request utils.ClientRegistry) error {
	exp, err := c.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create client registry record")
	}

	registration, err := c.store.GetClientRegistry(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}
	if *registration != (sqlstore.ClientRegistry{}) {
		return errors.New("client registry record already exists when create client registry record")
	}

	err = c.store.InsertClientRegistry(request)
	if err != nil {
		return err
	}

	return nil

}

func (s *ServerService) CreateClientSet(request utils.ClientSet) error {
	exp, err := s.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create server's client set record")
	}

	client_set, err := s.store.GetClientSet(request.Exp_ID, request.Server_ID)
	if err != nil {
		return err
	}
	if client_set.Exp_ID != "" && client_set.Server_ID != "" && len(client_set.Clients) != 0 {
		return errors.New("record already exists when create server's client set record")
	}

	err = s.store.InsertClientSet(request)
	if err != nil {
		return err
	}

	return nil

}

func (e *ExperimentService) CreateExp(request utils.OutputPartyRequest) error {
	exp, err := e.store.GetExp(request.Exp_ID)

	if err != nil {
		return err
	}

	if *exp != (sqlstore.Experiment{}) {
		return errors.New("experiment already exists when server creates experiment record")
	}

	err = e.store.InsertExp(request)

	if err != nil {
		return err
	}

	return nil
}
