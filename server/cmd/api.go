package main

import (
	"errors"
	"log"
	"time"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/server/config"
	"example.com/SMC/server/sqlstore"
)

type ClientService struct {
	db *sqlstore.DB
}

type ServerService struct {
	db *sqlstore.DB
}

type ExperimentService struct {
	db *sqlstore.DB
}

func NewClientService(db *sqlstore.DB) *ClientService {
	return &ClientService{db: db}
}

func NewServerService(db *sqlstore.DB) *ServerService {
	return &ServerService{db: db}
}

func NewExperimentService(db *sqlstore.DB) *ExperimentService {
	return &ExperimentService{db: db}
}

func (c *ClientService) CreateClientShare(request ClientRequest, cfg *config.Server) error {
	// TODO : do some basic validations, e.g. missing exp_id, client_id, etc.
	exp, err := c.db.GetExperiment(request.Exp_ID)
	if err != nil {
		return err
	}

	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when server creates client share")
	}

	//timestamp, _ := time.Parse("2006-01-02 15:04:05", request.Timestamp)
	//due, _ := time.Parse("2006-01-02 15:04:05", exp.ClientShareDue)
	timestamp := time.Now().UTC()
	due := timestamp.Add(1 * time.Minute)

	if timestamp.After(due) {
		return errors.New("client submitted share after due")
	}

	//insert experiment id and client id to client table
	err = c.db.InsertClient(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}

	//insert share to client share table
	for input_index, party := range request.Proof.PartyShares {
		for _, share := range party.Shares {
			err = c.db.InsertClientShare(request.Exp_ID, request.Client_ID, input_index, share.Index, share.Value)
			if err != nil {
				return err
			}
		}
	}

	zk, err := ligero.NewLigeroZK(cfg.N_secrets, cfg.M, cfg.N, cfg.T, cfg.Q, cfg.N_open)
	if err != nil {
		log.Fatal(err)
	}

	verify, err := zk.VerifyProof(request.Proof)

	//creat complaint record based on proof verification result
	if !verify {
		log.Printf("failed to verify %s proof for %s -- %s\n", request.Client_ID, request.Exp_ID, err)

		/**
		//test s6 should complaint but not complaint
		if cfg.Server_ID == "s6"{
			_ = c.db.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, false, request.Proof.MerkleRoot)
		} else {
			err = c.db.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, true, request.Proof.MerkleRoot)
			if err != nil {
				panic(err)
			}
		}**/

		err = c.db.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, true, request.Proof.MerkleRoot)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("succeed to verify %s proof for %s\n", request.Client_ID, request.Exp_ID)

		/**
		//test s6 should not complaint but complaint
		if cfg.Server_ID == "s6" && request.Client_ID == "c1" {
			_ = c.db.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, true, request.Proof.MerkleRoot)
		} else {
			err = c.db.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, false, request.Proof.MerkleRoot)
			if err != nil {
				panic(err)
			}
		}**/

		err = c.db.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, false, request.Proof.MerkleRoot)
		if err != nil {
			panic(err)
		}

	}

	return nil
}

func (s *ServerService) CreateComplaint(request ComplaintRequest) error {
	exp, err := s.db.GetExperiment(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create server's complaint")
	}

	for _, comp := range request.Complaints {
		err = s.db.InsertComplaint(request.Exp_ID, request.Server_ID, comp.Client_ID, comp.Complain, comp.Root)
		if err != nil {
			return err
		}
	}

	return nil

}

func (s *ServerService) CreateMaskedShares(request MaskedShareRequest) error {
	exp, err := s.db.GetExperiment(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create other servers' masked shares")
	}

	log.Printf("server received masked shares from %s: %+v\n", request.Server_ID, request)
	for _, masked_sh := range request.MaskedShares {
		err = s.db.InsertMaskedShare(request.Exp_ID, request.Server_ID, masked_sh.Client_ID, masked_sh.Input_Index, masked_sh.Index, masked_sh.Value)
		if err != nil {
			panic(err)
		}

	}

	return nil

}

func (s *ServerService) CreateValidClient(exp_id, client_id string) error {
	exp, err := s.db.GetExperiment(exp_id)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create valid client")
	}

	err = s.db.InsertValidClient(exp_id, client_id)
	if err != nil {
		return err
	}

	return nil

}

func (e *ExperimentService) CreateExperiment(request OutputPartyRequest) error {
	//TODO: need to remove and use due in the file
	complaint_due := time.Now().UTC().Add(time.Duration(4) * time.Minute).Format("2006-01-02 15:04:05")
	share_broadcast_due := time.Now().UTC().Add(time.Duration(7) * time.Minute).Format("2006-01-02 15:04:05")

	err := e.db.InsertExperiment(request.Exp_ID, request.ClientShareDue, complaint_due, share_broadcast_due, request.Owner)

	if err != nil {
		return err
	}

	return nil
}

func (c *ClientService) CreateClientRegistry(request ClientRegistry) error {
	exp, err := c.db.GetExperiment(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create client registry")
	}

	err = c.db.InsertClientRegistry(request.Exp_ID, request.Client_ID, request.Token)
	if err != nil {
		return err
	}

	return nil

}
