package sqlstore

import (
	"fmt"
	"log"

	"example.com/SMC/server/public/utils"
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

	db.AutoMigrate(&ClientShare{})

	db.AutoMigrate(&ClientRegistry{})

	db.AutoMigrate(&ClientSet{})

	db.AutoMigrate(&ServerComputation{})

	return db, nil
}

/**
func (storage *Storage) Migrate() {
	storage.db.AutoMigrate(&Experiment{})

	storage.db.AutoMigrate(&Client{})

	storage.db.AutoMigrate(&ClientRegistry{})

	storage.db.AutoMigrate(&Server{})
}**/

// create client share record in the client table
func (store *SqlStore) InsertClientShare(client utils.ClientRequest) error {
	c := ClientShare{
		Exp_ID:      client.Exp_ID,
		Client_ID:   client.Client_ID,
		Share_Index: client.Secret_Share.Index,
		Share_Value: client.Secret_Share.Value,
	}
	result := store.db.Create(&c)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// create client registration record in the client registration table
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
}

func (store *SqlStore) InsertClientSet(client_set utils.ClientSet) error {
	cs := ClientSet{
		Exp_ID:    client_set.Exp_ID,
		Server_ID: client_set.Server_ID,
		Clients:   client_set.Clients,
	}
	result := store.db.Create(&cs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// create server sumShare record in the server table
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
}

// create experiment record in the experiment tables
func (store *SqlStore) InsertExp(experiment utils.OutputPartyRequest) error {
	exp := &Experiment{
		Exp_ID:    experiment.Exp_ID,
		Due:       experiment.Due,
		Owner:     experiment.Owner,
		Completed: false,
	}
	result := store.db.Create(&exp)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// get client share record
func (store *SqlStore) GetClient(exp_id string, client_id string) (*ClientShare, error) {
	var client ClientShare
	r := store.db.Find(&client, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &client, nil
}

// get client registration record
func (store *SqlStore) GetClientRegistry(exp_id string, client_id string) (*ClientRegistry, error) {
	var cr ClientRegistry
	r := store.db.Find(&cr, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &cr, nil
}

// get a server's client set
func (store *SqlStore) GetClientSet(exp_id string, server_id string) (*ClientSet, error) {
	var cs ClientSet
	r := store.db.Find(&cs, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &cs, nil
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

// get server sumShare record
func (store *SqlStore) GetServerComputation(exp_id string, server_id string) (*ServerComputation, error) {
	var server ServerComputation
	r := store.db.Find(&server, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &server, nil
}

// get clients' share records of an experiment
func (store *SqlStore) GetAllClients(exp_id string) ([]ClientShare, error) {
	var clients []ClientShare
	r := store.db.Find(&clients, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return clients, nil
}

// get each server's client set of an experiment
func (store *SqlStore) GeAllClientSets(exp_id string) ([]ClientSet, error) {
	var client_sets []ClientSet
	r := store.db.Find(&client_sets, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return client_sets, nil
}

// get clients' registration records of an experiment
func (store *SqlStore) GetRegisteredClient(exp_id string) ([]ClientRegistry, error) {
	var registrations []ClientRegistry
	r := store.db.Find(&registrations, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return registrations, nil
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

// get all experiments records that server round is completed but sum share is not completed
func (store *SqlStore) GetAllExpsWithServerRoundCompleted() ([]Experiment, error) {
	var experiments []Experiment
	r := store.db.Find(&experiments, "server_round_completed = ? and completed=?", true, false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get servers' sumShare records of an experiment
func (store *SqlStore) GetAllServers(exp_id string) ([]ServerComputation, error) {
	var servers []ServerComputation
	r := store.db.Find(&servers, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}

	return servers, nil
}

// set experiment's server round to completed
func (store *SqlStore) UpdateHalfCompletedExperiment(exp_id string) error {
	var exp Experiment
	r := store.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Server_Round_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// set experiment status to completed
func (store *SqlStore) UpdateCompletedExperiment(exp_id string) error {
	var exp Experiment
	r := store.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Completed", true)
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

// delete server record from server table
func (store *SqlStore) DeleteServer(exp_id string) error {
	r := store.db.Delete(&ServerComputation{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete client registration record from client table
func (store *SqlStore) DeleteClientRegistry(exp_id string) error {
	r := store.db.Delete(&ClientRegistry{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}
