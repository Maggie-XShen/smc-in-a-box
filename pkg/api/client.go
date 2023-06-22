package api

import (
	"errors"
	"time"

	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/repository"
)

type ClientService struct {
	storage repository.Storage
}

func NewClientService(s repository.Storage) *ClientService {
	return &ClientService{storage: s}
}

func (c *ClientService) CreateClient(request message.ClientRequest) error {
	// TODO : do some basic validations, e.g. missing exp_id, client_id, etc.
	exp, err := c.storage.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}

	if *exp == (repository.Experiment{}) {
		return errors.New("experiment does not exist when create client share record")
	}

	timestamp, _ := time.Parse("2006-01-02 15:04:05", request.Timestamp)
	due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)
	if timestamp.After(due) {
		return errors.New("client share submission passed due when create client share record")
	}

	client, err := c.storage.GetClient(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}
	if *client != (repository.Client{}) {
		return errors.New("client record already exists create client share record")
	}

	//Todo: check bad events of client share

	err = c.storage.CreateClient(request)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClientService) CreateClientRegistry(request message.ClientRegistry) error {
	exp, err := c.storage.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (repository.Experiment{}) {
		return errors.New("experiment does not exist when create client registry record")
	}

	registration, err := c.storage.GetClientRegistry(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}
	if *registration != (repository.ClientRegistry{}) {
		return errors.New("client registry record already exists when create client registry record")
	}

	err = c.storage.CreateClientRegistration(request)
	if err != nil {
		return err
	}

	return nil

}
