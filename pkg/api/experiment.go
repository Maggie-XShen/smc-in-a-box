package api

import (
	"errors"

	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/repository"
)

type ExperimentService struct {
	storage repository.Storage
}

func NewExperimentService(s repository.Storage) *ExperimentService {
	return &ExperimentService{storage: s}
}

func (e *ExperimentService) CreateExp(request message.OutputPartyRequest) error {
	exp, err := e.storage.GetExp(request.Exp_ID)

	if err != nil {
		return err
	}

	if *exp != (repository.Experiment{}) {
		return errors.New("experiment already exists")
	}

	err = e.storage.CreateExp(request)
	if err != nil {
		return err
	}

	return nil
}
