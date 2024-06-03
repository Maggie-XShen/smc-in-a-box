package sqlstore

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	DB *gorm.DB
}

func NewDB(id string) *DB {
	db, err := SetupDatabase(id)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}
	return &DB{DB: db}

}

func SetupDatabase(sid string) (*gorm.DB, error) {
	//dsn := fmt.Sprintf("smc:smcinabox@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", sid)
	dsn := "smc:smcinabox@tcp(127.0.0.1:3306)/smc?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Nanosecond, // Set the threshold to a very low value
			LogLevel:      logger.Silent,   // Set log level to Silent
			Colorful:      false,           // Disable color
		},
	)

	// Open a connection to the MySQL database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}

	log.Printf("Connection to %s Database Established\n", sid)

	// Auto-migrate tables
	if err := db.AutoMigrate(&Experiment{}, &Client{}, &ClientShare{}, &Complaint{}, &ValidClient{}, &MaskedShare{}); err != nil {
		return nil, err
	}

	return db, nil
}

func DeleteDB(db_name string) {
	// remove old database
	os.Remove(db_name)
}

// create client table
func (db *DB) InsertClient(exp_id, client_id string) error {
	c := Client{
		Exp_ID:    exp_id,
		Client_ID: client_id,
	}
	result := db.DB.Create(&c)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get all clients per experiments
func (db *DB) GetClientsPerExperiment(exp_id string) ([]Client, error) {
	var client []Client
	r := db.DB.Find(&client, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return client, nil
}

// create client share record
func (db *DB) InsertClientShare(exp_id, client_id string, shares []byte) error {
	c := ClientShare{
		Exp_ID:    exp_id,
		Client_ID: client_id,
		Shares:    shares,
	}
	result := db.DB.Create(&c)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get client share record
func (db *DB) GetClientShares(exp_id string, client_id string) (ClientShare, error) {
	var client ClientShare
	r := db.DB.Find(&client, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return ClientShare{}, r.Error
	}
	return client, nil
}

// get clients' share records of an experiment
func (db *DB) GetClientsSharesPerExperiment(exp_id string) ([]ClientShare, error) {
	var clients []ClientShare
	r := db.DB.Find(&clients, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return clients, nil
}

// update client's share
func (db *DB) UpdateClientShare(exp_id, client_id string, shares []byte) error {
	/**
	r := db.db.Model(&ClientShare{}).Where("exp_id = ? and client_id = ? and input_index = ? and index = ?", exp_id, client_id, input_index, index).Update("value", value)
	if r.Error != nil {
		return r.Error
	}**/
	db.DB.Save(&ClientShare{Exp_ID: exp_id, Client_ID: client_id, Shares: shares})
	return nil
}

// get valid client's share
func (db *DB) GetValidClientShares(exp_id string) ([]ClientShare, error) {
	var clients []ClientShare

	/**
	r := db.db.Model(&ValidClient{}).Joins("INNER JOIN client_shares ON validclient.exp_id = clientshare.exp_id and validclient.client_id = clientshare.client_id").
		Select("clientshare.exp_id, clientshare.client_id, clientshare.input_index, clientshare.index, clientshare.value").Where("validclient.exp_id = ?", exp_id).
		Find(&clients)
	if r.Error != nil {
		return nil, r.Error
	}**/
	valid_clients, err := db.GetValidClientsPerExperiment(exp_id)
	if err != nil {
		return nil, err
	}

	for _, vc := range valid_clients {
		shares, err := db.GetClientShares(exp_id, vc.Client_ID)
		if err != nil {
			return nil, err
		}

		clients = append(clients, shares)

	}
	return clients, nil
}

func (db *DB) InsertComplaint(exp_id, server_id, client_id string, isComplain bool, mkt_root []byte) error {
	comp := Complaint{
		Exp_ID:    exp_id,
		Server_ID: server_id,
		Client_ID: client_id,
		Complain:  isComplain,
		Root:      mkt_root,
	}
	result := db.DB.Create(&comp)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get all complaints of an experiment
func (db *DB) GetComplaintsPerExperiment(exp_id string) ([]Complaint, error) {
	var comp []Complaint
	r := db.DB.Find(&comp, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return comp, nil
}

// get count of complaints of an experiment
func (db *DB) CountComplaintsPerExperiment(exp_id string) int64 {
	var count int64
	db.DB.Model(&Complaint{}).Where("exp_id = ?", exp_id).Count(&count)
	return count
}

// get a complaint record
func (db *DB) GetComplaint(exp_id, server_id, client_id string) (*Complaint, error) {
	var comp Complaint
	r := db.DB.Find(&comp, "exp_id = ? and server_id = ? and client_id = ?", exp_id, server_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &comp, nil
}

// get complaint record where complaint is false
func (db *DB) GetNoComplain(exp_id, client_id string) ([]Complaint, error) {
	var comp []Complaint
	r := db.DB.Find(&comp, "exp_id = ? and client_id = ? and complain=?", exp_id, client_id, false)
	if r.Error != nil {
		return nil, r.Error
	}
	return comp, nil
}

// get clients in complaint table but not in client table
func (db *DB) GetDropoutClient(exp_id string) ([]string, error) {
	var client []struct {
		Client_ID string
	}
	sub := db.DB.Model(&Client{}).Select("client_id").Where("exp_id = ?", exp_id)

	r := db.DB.Model(&Complaint{}).Select("client_id").Where("client_id NOT IN (?)", sub).Group("client_id").Find(&client)
	if r.Error != nil {
		return nil, r.Error
	}

	var result []string
	for _, c := range client {
		result = append(result, c.Client_ID)
	}

	return result, nil
}

// get complaint records per exp_id, server_id
func (db *DB) GetComplaintsPerServer(exp_id, server_id string) ([]Complaint, error) {
	var comp []Complaint
	r := db.DB.Find(&comp, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return comp, nil
}

// get complaint records per exp_id, client_id
func (db *DB) GetComplaintsPerClient(exp_id, client_id string) ([]Complaint, error) {
	var comp []Complaint
	r := db.DB.Find(&comp, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return comp, nil
}

func (db *DB) InsertEchoComplaint(exp_id, server_id, complaints string) error {
	echo := EchoComplaint{
		Exp_ID:     exp_id,
		Server_ID:  server_id,
		Complaints: complaints,
	}
	result := db.DB.Create(&echo)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get echo complaint record
func (db *DB) GetEchoComplaint(exp_id, server_id, complaints string) ([]EchoComplaint, error) {
	var echo []EchoComplaint
	r := db.DB.Find(&echo, "exp_id = ? and server_id = ? and complaints = ?", exp_id, server_id, complaints)
	if r.Error != nil {
		return nil, r.Error
	}
	return echo, nil
}

func (db *DB) InsertValidClient(exp_id, client_id string) error {
	vc := ValidClient{
		Exp_ID:    exp_id,
		Client_ID: client_id,
	}
	result := db.DB.Create(&vc)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get all valid clients
func (db *DB) GetValidClientsPerExperiment(exp_id string) ([]ValidClient, error) {
	var vc []ValidClient
	r := db.DB.Find(&vc, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return vc, nil
}

// delete a client from valid client table
func (db *DB) DeleteValidClient(exp_id, client_id string) error {
	r := db.DB.Delete(&ValidClient{Exp_ID: exp_id, Client_ID: client_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func (db *DB) InsertMaskedShare(exp_id, server_id, client_id string, shares []byte) error {
	mask_share := MaskedShare{
		Exp_ID:    exp_id,
		Server_ID: server_id,
		Client_ID: client_id,
		Shares:    shares,
	}
	result := db.DB.Create(&mask_share)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db *DB) GetMaskedSharesPerClient(exp_id, server_id, client_id string) (MaskedShare, error) {
	var masked_shares MaskedShare
	r := db.DB.Find(&masked_shares, "exp_id = ? and server_id = ? and client_id = ?", exp_id, server_id, client_id)
	if r.Error != nil {
		return MaskedShare{}, r.Error
	}
	return masked_shares, nil
}

func (db *DB) GetMaskedSharesPerServer(exp_id, server_id string) ([]MaskedShare, error) {
	var masked_shares []MaskedShare
	r := db.DB.Find(&masked_shares, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return masked_shares, nil
}

func (db *DB) GetMaskedSharesPerExperiment(exp_id string) ([]MaskedShare, error) {
	var masked_shares []MaskedShare
	r := db.DB.Find(&masked_shares, "exp_id = ? ", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return masked_shares, nil
}

func (db *DB) CountMaskedSharesPerExperiment(exp_id string) int64 {
	var count int64
	db.DB.Model(&MaskedShare{}).Where("exp_id = ?", exp_id).Count(&count)
	return count
}

func (db *DB) InsertEchoMaskedShare(exp_id, server_id, mask_shares string) error {
	echo := EchoMaskedShare{
		Exp_ID:       exp_id,
		Server_ID:    server_id,
		MaskedShares: mask_shares,
	}
	result := db.DB.Create(&echo)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get echo masked share record
func (db *DB) GetEchoMaskedShare(exp_id, server_id, mask_shares string) ([]EchoMaskedShare, error) {
	var echo []EchoMaskedShare
	r := db.DB.Find(&echo, "exp_id = ? and server_id = ? and maskedshares = ?", exp_id, server_id, mask_shares)
	if r.Error != nil {
		return nil, r.Error
	}
	return echo, nil
}

// create experiment record in the experiment tables
func (db *DB) InsertExperiment(exp_id, due1, due2, due3, owner string) error {
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
	result := db.DB.Create(&exp)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// get experiment record
func (db *DB) GetExperiment(exp_id string) (*Experiment, error) {
	var exp Experiment
	r := db.DB.Find(&exp, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &exp, nil
}

// get all experiments not pass client share submission due
func (db *DB) GetAllExperiments() ([]Experiment, error) {
	var experiments []Experiment
	r := db.DB.Find(&experiments, "Round1_Completed = ?", false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get experiments count
func (db *DB) GetExperimentCount() (int64, error) {
	var count int64
	err := db.DB.Model(&Experiment{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// get all experiments that round1 is completed
func (db *DB) GetExpsWithRound1Completed() ([]Experiment, error) {
	var experiments []Experiment
	r := db.DB.Find(&experiments, "round1_completed = ? and round2_completed=? and round3_completed=?", true, false, false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get all experiments records that round2 is completed
func (db *DB) GetExpsWithRound2Completed() ([]Experiment, error) {
	var experiments []Experiment
	r := db.DB.Find(&experiments, "round1_completed = ? and round2_completed=? and round3_completed=?", true, true, false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get all experiments records that round3 is completed
func (db *DB) GetExpsWithRound3Completed() ([]Experiment, error) {
	var experiments []Experiment
	r := db.DB.Find(&experiments, "round1_completed = ? and round2_completed=? and round3_completed=?", true, true, true)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// set client share submission round to completed
func (db *DB) UpdateRound1Completed(exp_id string) error {
	r := db.DB.Model(&Experiment{}).Where("exp_ID = ?", exp_id).Update("Round1_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// set complant broadcast round to completed
func (db *DB) UpdateRound2Completed(exp_id string) error {
	r := db.DB.Model(&Experiment{}).Where("exp_ID = ?", exp_id).Update("Round2_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// set experiment status to completed
func (db *DB) UpdateRound3Completed(exp_id string) error {
	r := db.DB.Model(&Experiment{}).Where("exp_ID = ?", exp_id).Update("Round3_Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete experiment record from experiment table
func (db *DB) DeleteExperiment(exp_id string) error {
	r := db.DB.Delete(&Experiment{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete client record from client table
func (db *DB) DeleteClient(exp_id string) error {
	r := db.DB.Delete(&ClientShare{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// create client registration record in the client registration table
func (db *DB) InsertClientRegistry(exp_id, client_id, token string) error {
	cr := ClientRegistry{
		Exp_ID:    exp_id,
		Client_ID: client_id,
		Token:     token,
	}
	result := db.DB.Create(&cr)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
