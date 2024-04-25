package sqlstore

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	db *gorm.DB
}

func NewDB(id string) *DB {
	db, err := SetupDatabase(id)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}
	return &DB{db: db}

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

	// Set the journal mode to WAL
	if err := db.Exec("PRAGMA journal_mode = WAL").Error; err != nil {
		return nil, err
	}

	db.AutoMigrate(&Experiment{})

	db.AutoMigrate(&ServerShare{})

	return db, nil
}

// create server sumShare record in the server table
func (db *DB) InsertServerShare(exp_id, server_id string, input_index, index, value int) error {
	s := ServerShare{
		Exp_ID:      exp_id,
		Server_ID:   server_id,
		Input_Index: input_index,
		Index:       index,
		Value:       value,
	}
	result := db.db.Clauses(clause.Insert{Modifier: "OR IGNORE"}).Create(&s)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get server's aggregated share
func (db *DB) GetSharesPerServer(exp_id, server_id string) ([]ServerShare, error) {
	var shares []ServerShare
	r := db.db.Find(&shares, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return shares, nil
}

// get all shares of an experiment
func (db *DB) GetSharesPerExperiment(exp_id string) ([]ServerShare, error) {
	var shares []ServerShare
	r := db.db.Find(&shares, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}

	return shares, nil
}

func (db *DB) CountSharesPerExperiment(exp_id string) int64 {
	var count int64
	db.db.Model(&ServerShare{}).Where("exp_id = ?", exp_id).Count(&count)

	return count
}

// create experiment record in the experiment tables
func (db *DB) InsertExperiment(exp_id, due1, due2 string) error {
	exp := &Experiment{
		Exp_ID:         exp_id,
		ClientShareDue: due1,
		ServerShareDue: due2,
		Completed:      false,
	}
	result := db.db.Clauses(clause.Insert{Modifier: "OR IGNORE"}).Create(&exp)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// get experiment record
func (db *DB) GetExperiment(exp_id string) (*Experiment, error) {
	var exp Experiment
	r := db.db.Find(&exp, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &exp, nil
}

// get all experiments records that server round is not completed
func (db *DB) GetAllExperiments() ([]Experiment, error) {
	var experiments []Experiment
	r := db.db.Find(&experiments, "completed=?", false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// set experiment status to completed
func (db *DB) UpdateCompletedExperiment(exp_id string) error {
	var exp Experiment
	r := db.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete experiment record from experiment table
func (db *DB) DeleteExperiment(exp_id string) error {
	r := db.db.Delete(&Experiment{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}
