package api

import (
	"errors"

	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/repository"
)

type ServerService struct {
	storage repository.Storage
}

func NewServerService(storage repository.Storage) *ServerService {
	return &ServerService{storage: storage}
}

func (ss *ServerService) CreateServer(request message.ServerRequest) error {
	//Todo: check validity of server

	exp, err := ss.storage.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (repository.Experiment{}) {
		return errors.New("experiment does not exist")
	}

	server, err := ss.storage.GetServer(request.Exp_ID, request.Server_ID)
	if err != nil {
		return err
	}
	if *server != (repository.Server{}) {
		return errors.New("server record already exists")
	}
	err = ss.storage.CreateServer(request)
	if err != nil {
		return err
	}

	return nil

}
