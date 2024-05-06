package main

import (
	"errors"
	"log"
	"time"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/server/config"
	"example.com/SMC/server/sqlstore"
	"github.com/sirupsen/logrus"
)

var client_count = 0

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

	timestamp, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", request.Timestamp)
	due, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", exp.ClientShareDue)

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

	proof_verify_start := time.Now() //proof verification start time
	verify, err := zk.VerifyProof(request.Proof)
	proof_verify_end := time.Since(proof_verify_start) //proof verification computing time
	logger.WithFields(logrus.Fields{
		"exp_id":      request.Exp_ID,
		"client_id":   request.Client_ID,
		"verify_time": proof_verify_end.String(),
		"is_verified": verify,
	}).Info("")
	total_verify_time += proof_verify_end
	//creat complaint record based on proof verification result
	if !verify {
		log.Printf("%s failed to verify %s proof for %s -- %s\n", cfg.Server_ID, request.Client_ID, request.Exp_ID, err)

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
		log.Printf("%s succeed to verify %s proof for %s\n", cfg.Server_ID, request.Client_ID, request.Exp_ID)

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

	log.Printf("server received masked shares from %s\n", request.Server_ID)
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

func (e *ExperimentService) CreateExperiment(request Experiment) error {
	err := e.db.InsertExperiment(request.Exp_ID, request.ClientShareDue, request.ComplaintDue, request.ShareBroadcastDue, request.Owner)

	if err != nil {
		return err
	}

	return nil
}

/**
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

}**/
