package main

import (
	"errors"
	"fmt"
	"log"
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

func (c *ClientService) CreateClientShare(request utils.ClientRequest, cfg *config.Server) error {
	// TODO : do some basic validations, e.g. missing exp_id, client_id, etc.
	exp, err := c.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}

	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when server creates client share record")
	}

	timestamp, _ := time.Parse("2006-01-02 15:04:05", request.Timestamp)
	due, _ := time.Parse("2006-01-02 15:04:05", exp.ClientShareDue)
	if timestamp.After(due) {
		return errors.New("client share submission passed due when server creates client shares record")
	}

	client, err := c.store.GetClient(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}
	if *client != (sqlstore.Client{}) {
		return errors.New("client record already exists when server creates client shares record")
	}

	//insert experiment id and client id to client table
	err = c.store.InsertClient(request.Exp_ID, request.Client_ID)
	if err != nil {
		return err
	}

	//insert share to client share table
	for input_index, party := range request.Proof.PartyShares {
		for _, share := range party.Shares {
			err = c.store.InsertClientShare(request.Exp_ID, request.Client_ID, input_index, share.Index, share.Value)
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
		log.Printf("failed to verify proof from %s for %s data\n", request.Client_ID, request.Exp_ID)
		fmt.Println(err)

		err = c.store.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, true, request.Proof.MerkleRoot)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("succeed to verify proof from %s for %s data\n", request.Client_ID, request.Exp_ID)

		err = c.store.InsertComplaint(request.Exp_ID, cfg.Server_ID, request.Client_ID, false, request.Proof.MerkleRoot)
		if err != nil {
			log.Fatal(err)
		}

	}

	return nil
}

/**
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

}**/

func (s *ServerService) CreateComplaint(request utils.ComplaintRequest) error {
	exp, err := s.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create server's complaint record")
	}

	//check if server's complaint records already exist
	complaint, err := s.store.GetAllComplaintsPerServer(request.Exp_ID, request.Server_ID)
	if err != nil {
		return err
	}
	if len(complaint) > 0 {
		return errors.New("record already exists when create server's complaint record")
	}

	for _, comp := range request.Complaints {
		err = s.store.InsertComplaint(request.Exp_ID, request.Server_ID, comp.Client_ID, comp.Complain, comp.Root)
		if err != nil {
			return err
		}
	}

	return nil

}

func (s *ServerService) CreateMaskedShares(request utils.MaskedShareRequest) error {
	exp, err := s.store.GetExp(request.Exp_ID)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create server's masked shares")
	}

	//check if server's masked shares already exist
	masked_shares, err := s.store.GetMaskedSharesPerServer(request.Exp_ID, request.Server_ID)
	if err != nil {
		return err
	}
	if len(masked_shares) > 0 {
		return errors.New("record already exists when create server's masked shares")
	}

	for _, masked_sh := range request.MaskedShares {
		err = s.store.InsertMaskedShare(request.Exp_ID, request.Server_ID, masked_sh.Client_ID, masked_sh.Input_Index, masked_sh.Index, masked_sh.Value)
		if err != nil {
			return err
		}

	}

	return nil

}

/**
func (s *ServerService) CreateMask(exp_id, client_id string, input_index, index, value int) error {
	exp, err := s.store.GetExp(exp_id)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create vss shares")
	}

	vss_shares, err := s.store.GetMask(exp_id, client_id, input_index)
	if err != nil {
		return err
	}
	if len(vss_shares) > 0 {
		return errors.New("record already exists when create vss shares")
	}

	err = s.store.InsertMask(exp_id, client_id, input_index, index, value)
	if err != nil {
		return err
	}

	return nil

}**/

func (s *ServerService) CreateValidClient(exp_id, client_id string) error {
	exp, err := s.store.GetExp(exp_id)
	if err != nil {
		return err
	}
	if *exp == (sqlstore.Experiment{}) {
		return errors.New("experiment does not exist when create valid client record")
	}

	client_set, err := s.store.GetValidClient(exp_id, client_id)
	if err != nil {
		return err
	}
	if client_set.Exp_ID != "" && client_set.Client_ID != "" {
		return errors.New("record already exists when create valid client record")
	}

	err = s.store.InsertValidClient(exp_id, client_id)
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

	complaint_due := time.Now().UTC().Add(time.Duration(4) * time.Minute).Format("2006-01-02 15:04:05")
	share_broadcast_due := time.Now().UTC().Add(time.Duration(7) * time.Minute).Format("2006-01-02 15:04:05")
	err = e.store.InsertExp(request.Exp_ID, request.ClientShareDue, complaint_due, share_broadcast_due, request.Owner)

	if err != nil {
		return err
	}

	return nil
}
