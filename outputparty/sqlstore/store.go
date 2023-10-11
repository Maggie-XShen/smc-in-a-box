package sqlstore

import (
	"fmt"
	"log"

	"example.com/SMC/outputparty/public/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlStore struct {
	db     *gorm.DB
	stores SqlStoreStores
}

type SqlStoreStores struct {
	serverComputationStore ServerComputation
	experimentStatusStore  Experiment
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

// delete server record from server table
func (store *SqlStore) DeleteServer(exp_id string) error {
	r := store.db.Delete(&ServerComputation{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}
