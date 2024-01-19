package sqlstore

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlStore struct {
	db *gorm.DB
}

func New(id string) *SqlStore {
	db, err := SetupDatabase(id)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}
	return &SqlStore{db: db}

}

func SetupDatabase(sid string) (*gorm.DB, error) {
	db_path := fmt.Sprintf("%s.db", sid)

	// remove old database
	//os.Remove(db_name)

	// open a database
	db, err := gorm.Open(sqlite.Open(db_path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Printf("Connection to %s Database Established\n", sid)

	db.AutoMigrate(&Experiment{})

	db.AutoMigrate(&Client{})

	db.AutoMigrate(&ClientShare{})

	db.AutoMigrate(&Complaint{})

	db.AutoMigrate(&ValidClient{})

	db.AutoMigrate(&Mask{})

	db.AutoMigrate(&MaskedShare{})

	return db, nil
}

// create client table
func (store *SqlStore) InsertClient(exp_id, client_id string) error {
	c := Client{
		Exp_ID:    exp_id,
		Client_ID: client_id,
	}
	result := store.db.Create(&c)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get client record
func (store *SqlStore) GetClient(exp_id, client_id string) (*Client, error) {
	var client Client
	r := store.db.Find(&client, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &client, nil
}

// get all clients per experiments
func (store *SqlStore) GetAllClients(exp_id string) ([]Client, error) {
	var client []Client
	r := store.db.Find(&client, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return client, nil
}

// create client share record
func (store *SqlStore) InsertClientShare(exp_id, client_id string, input_index, sh_index, sh_value int) error {
	c := ClientShare{
		Exp_ID:      exp_id,
		Client_ID:   client_id,
		Input_Index: input_index,
		Share_Index: sh_index,
		Share_Value: sh_value,
	}
	result := store.db.Create(&c)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get client share record
func (store *SqlStore) GetClientShares(exp_id string, client_id string) ([]ClientShare, error) {
	var client []ClientShare
	r := store.db.Find(&client, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return client, nil
}

// get clients' share records of an experiment
func (store *SqlStore) GetAllClientsShares(exp_id string) ([]ClientShare, error) {
	var clients []ClientShare
	r := store.db.Find(&clients, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return clients, nil
}

// get clients' share records of an experiment
func (store *SqlStore) GetValidClientShares(exp_id string) ([]ClientShare, error) {
	var clients []ClientShare
	r := store.db.Table("validclient").Joins("INNER JOIN clientshare ON validclient.exp_id = clientshare.exp_id and validclient.client_id = clientshare.client_id").
		Select("clientshare.exp_id, clientshare.client_id, clientshare.input_index, clientshare.share_index, clientshare.share_value").Where("validclient.exp_id = ?", exp_id).
		Find(&clients)
	if r.Error != nil {
		return nil, r.Error
	}
	return clients, nil
}

func (store *SqlStore) InsertComplaint(exp_id, server_id, client_id string, isComplain bool, mkt_root []byte) error {
	comp := Complaint{
		Exp_ID:    exp_id,
		Server_ID: server_id,
		Client_ID: client_id,
		Complain:  isComplain,
		Root:      mkt_root,
	}
	result := store.db.Create(&comp)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get a complaint record
func (store *SqlStore) GetComplaint(exp_id, server_id, client_id string) (*Complaint, error) {
	var comp Complaint
	r := store.db.Find(&comp, "exp_id = ? and server_id = ? and client_id", exp_id, server_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &comp, nil
}

// get complaint records per exp_id, server_id
func (store *SqlStore) GetAllComplaintsPerServer(exp_id, server_id string) ([]Complaint, error) {
	var comp []Complaint
	r := store.db.Find(&comp, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return comp, nil
}

// get complaint records per exp_id, client_id
func (store *SqlStore) GetAllComplaintsPerClient(exp_id, client_id string) ([]Complaint, error) {
	var comp []Complaint
	r := store.db.Find(&comp, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return comp, nil
}

func (store *SqlStore) InsertValidClient(exp_id, client_id string) error {
	vc := ValidClient{
		Exp_ID:    exp_id,
		Client_ID: client_id,
	}
	result := store.db.Create(&vc)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get valid client
func (store *SqlStore) GetValidClient(exp_id, client_id string) (*ValidClient, error) {
	var vc ValidClient
	r := store.db.Find(&vc, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &vc, nil
}

// get all valid clients
func (store *SqlStore) GetAllValidClients(exp_id string) ([]ValidClient, error) {
	var vc []ValidClient
	r := store.db.Find(&vc, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return vc, nil
}

func (store *SqlStore) InsertMask(exp_id, client_id string, input_index, index, value int) error {
	vss := Mask{
		Exp_ID:      exp_id,
		Client_ID:   client_id,
		Input_Index: input_index,
		Index:       index,
		Value:       value,
	}
	result := store.db.Create(&vss)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (store *SqlStore) GetMask(exp_id, client_id string, input_index int) ([]Mask, error) {
	var vss []Mask
	r := store.db.Find(&vss, "exp_id = ? and client_id = ? and input_index", exp_id, client_id, input_index)
	if r.Error != nil {
		return nil, r.Error
	}
	return vss, nil
}

func (store *SqlStore) InsertMaskedShare(exp_id, server_id, client_id string, input_index, index, value int) error {
	mask_share := MaskedShare{
		Exp_ID:      exp_id,
		Server_ID:   server_id,
		Client_ID:   client_id,
		Input_Index: input_index,
		Index:       index,
		Value:       value,
	}
	result := store.db.Create(&mask_share)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (store *SqlStore) GetMaskedShares(exp_id, server_id, client_id string) ([]MaskedShare, error) {
	var masked_shares []MaskedShare
	r := store.db.Find(&masked_shares, "exp_id = ? and server_id = ? and client_id = ?", exp_id, server_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return masked_shares, nil
}

func (store *SqlStore) GetMaskedSharesPerServer(exp_id, server_id string) ([]MaskedShare, error) {
	var masked_shares []MaskedShare
	r := store.db.Find(&masked_shares, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return masked_shares, nil
}

func (store *SqlStore) GetMaskedSharesPerExp(exp_id string) ([]MaskedShare, error) {
	var masked_shares []MaskedShare
	r := store.db.Find(&masked_shares, "exp_id = ? ", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return masked_shares, nil
}

// create experiment record in the experiment tables
func (store *SqlStore) InsertExp(exp_id, due1, due2, due3, owner string) error {
	exp := &Experiment{
		Exp_ID:            exp_id,
		ClientShareDue:    due1,
		ComplaintDue:      due2,
		ShareBroadcastDue: due3,
		Owner:             owner,
		Round1_Completed:  false,
		Round2_Completed:  false,
		Round3_Completed:  false,
	}
	result := store.db.Create(&exp)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// get experiment record
func (store *SqlStore) GetExp(exp_id string) (*Experiment, error) {
	var exp Experiment
	r := store.db.Find(&exp, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &exp, nil
}

// get all experiments records that server round is not completed
func (store *SqlStore) GetAllExps() ([]Experiment, error) {
	var experiments []Experiment
	r := store.db.Find(&experiments, "server_round_completed = ? and completed=?", false, false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get all experiments records that round1 is completed
func (store *SqlStore) GetExpsWithSRound1Completed() ([]Experiment, error) {
	var experiments []Experiment
	r := store.db.Find(&experiments, "round1_completed = ? and round2_completed=? and round3_completed=?", true, false, false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get all experiments records that round2 is completed
func (store *SqlStore) GetExpsWithSRound2Completed() ([]Experiment, error) {
	var experiments []Experiment
	r := store.db.Find(&experiments, "round1_completed = ? and round2_completed=? and round3_completed=?", true, true, false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// set client share submission round to completed
func (store *SqlStore) UpdateRound1Completed(exp_id string) error {
	var exp Experiment
	r := store.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Round1_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// set complant broadcast round to completed
func (store *SqlStore) UpdateRound2Completed(exp_id string) error {
	var exp Experiment
	r := store.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Round2_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// set experiment status to completed
func (store *SqlStore) UpdateRound3Completed(exp_id string) error {
	var exp Experiment
	r := store.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Round3_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete experiment record from experiment table
func (store *SqlStore) DeleteExperiment(exp_id string) error {
	r := store.db.Delete(&Experiment{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete client record from client table
func (store *SqlStore) DeleteClient(exp_id string) error {
	r := store.db.Delete(&ClientShare{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// create client registration record in the client registration table
/**
func (store *SqlStore) InsertClientRegistry(reg utils.ClientRegistry) error {
	cr := ClientRegistry{
		Exp_ID:    reg.Exp_ID,
		Client_ID: reg.Client_ID,
		Token:     reg.Token,
	}
	result := store.db.Create(&cr)
	if result.Error != nil {
		return result.Error
	}
	return nil
}**/

// create server sumShare record in the server table
/**
func (store *SqlStore) InsertServerComputation(server utils.ServerRequest) error {
	s := ServerComputation{
		Exp_ID:         server.Exp_ID,
		Server_ID:      server.Server_ID,
		SumShare_Value: server.Sum_Shares.Value,
		SumShare_Index: server.Sum_Shares.Index,
	}
	result := store.db.Create(&s)
	if result.Error != nil {
		return result.Error
	}
	return nil
}**/

// get client registration record
/**
func (store *SqlStore) GetClientRegistry(exp_id string, client_id string) (*ClientRegistry, error) {
	var cr ClientRegistry
	r := store.db.Find(&cr, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &cr, nil
}**/

// get server sumShare record
/**
func (store *SqlStore) GetServerComputation(exp_id string, server_id string) (*ServerComputation, error) {
	var server ServerComputation
	r := store.db.Find(&server, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &server, nil
}**/

// get clients' registration records of an experiment
/**
func (store *SqlStore) GetRegisteredClient(exp_id string) ([]ClientRegistry, error) {
	var registrations []ClientRegistry
	r := store.db.Find(&registrations, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return registrations, nil
}**/

// get servers' sumShare records of an experiment
/**
func (store *SqlStore) GetAllServers(exp_id string) ([]ServerComputation, error) {
	var servers []ServerComputation
	r := store.db.Find(&servers, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}

	return servers, nil
}**/

// delete server record from server table
/**
func (store *SqlStore) DeleteServer(exp_id string) error {
	r := store.db.Delete(&ServerComputation{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}**/

// delete client registration record from client table
/**
func (store *SqlStore) DeleteClientRegistry(exp_id string) error {
	r := store.db.Delete(&ClientRegistry{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}**/
