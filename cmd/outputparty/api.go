package main

import (
	"errors"

	"example.com/SMC/pkg/repository"
	"example.com/SMC/pkg/utils/message"
)

type ServerService struct {
	storage *repository.Storage
}

func NewServerService(storage *repository.Storage) *ServerService {
	return &ServerService{storage: storage}
}

func (ss *ServerService) CreateServer(request message.ServerRequest) error {
	//Todo: check validity of server

	exp, err := ss.storage.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (repository.Experiment{}) {
		return errors.New("experiment does not exist when output party creates server sumshare record")
	}

	server, err := ss.storage.GetServer(request.Exp_ID, request.Server_ID)
	if err != nil {
		return err
	}
	if *server != (repository.Server{}) {
		return errors.New("server record already exists when output party creates server sumshare record")
	}
	err = ss.storage.CreateServer(request)
	if err != nil {
		return err
	}

	return nil

}

type ExperimentService struct {
	storage *repository.Storage
}

func NewExperimentService(s *repository.Storage) *ExperimentService {
	return &ExperimentService{storage: s}
}

func (e *ExperimentService) CreateExp(request message.OutputPartyRequest) error {
	exp, err := e.storage.GetExp(request.Exp_ID)

	if err != nil {
		return err
	}

	if *exp != (repository.Experiment{}) {
		return errors.New("experiment already exists when output party creates experiment record")
	}

	err = e.storage.CreateExp(request)
	if err != nil {
		return err
	}

	return nil
}
