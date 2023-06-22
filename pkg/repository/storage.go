package repository

import (
	"example.com/SMC/pkg/message"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

type Client struct {
	Exp_ID      string
	Client_ID   string
	Share_Index int
	Share_Value int
}

type ClientRegistry struct {
	Exp_ID    string
	Client_ID string
	Token     string
}

type Experiment struct {
	Exp_ID    string
	Due       string
	Completed bool
}

type Server struct {
	Exp_ID         string
	Server_ID      string
	SumShare_Value int
	SumShare_Index int
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{db: db}

}

func (storage *Storage) Migrate() {
	storage.db.AutoMigrate(&Experiment{})

	storage.db.AutoMigrate(&Client{})

	storage.db.AutoMigrate(&ClientRegistry{})

	storage.db.AutoMigrate(&Server{})
}

// create client share record in the client table
func (storage *Storage) CreateClient(client message.ClientRequest) error {
	c := Client{
		Exp_ID:      client.Exp_ID,
		Client_ID:   client.Client_ID,
		Share_Index: client.Secret_Share.Index,
		Share_Value: client.Secret_Share.Value,
	}
	result := storage.db.Create(&c)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// create client registration record in the client registration table
func (storage *Storage) CreateClientRegistration(reg message.ClientRegistry) error {
	cr := ClientRegistry{
		Exp_ID:    reg.Exp_ID,
		Client_ID: reg.Client_ID,
		Token:     reg.Token,
	}
	result := storage.db.Create(&cr)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// create server sumShare record in the server table
func (storage *Storage) CreateServer(server message.ServerRequest) error {
	s := Server{
		Exp_ID:         server.Exp_ID,
		Server_ID:      server.Server_ID,
		SumShare_Value: server.Sum_Shares.Value,
		SumShare_Index: server.Sum_Shares.Index,
	}
	result := storage.db.Create(&s)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// create experiment record in the experiment tables
func (storage *Storage) CreateExp(experiment message.OutputPartyRequest) error {
	exp := &Experiment{
		Exp_ID:    experiment.Exp_ID,
		Due:       experiment.Due,
		Completed: false,
	}
	result := storage.db.Create(&exp)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get client share record
func (storage *Storage) GetClient(exp_id string, client_id string) (*Client, error) {
	var client Client
	r := storage.db.Find(&client, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &client, nil
}

// get client registration record
func (storage *Storage) GetClientRegistry(exp_id string, client_id string) (*ClientRegistry, error) {
	var cr ClientRegistry
	r := storage.db.Find(&cr, "exp_id = ? and client_id = ?", exp_id, client_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &cr, nil
}

// get experiment record
func (storage *Storage) GetExp(exp_id string) (*Experiment, error) {
	var exp Experiment
	r := storage.db.Find(&exp, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &exp, nil
}

// get server sumShare record
func (storage *Storage) GetServer(exp_id string, server_id string) (*Server, error) {
	var server Server
	r := storage.db.Find(&server, "exp_id = ? and server_id = ?", exp_id, server_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return &server, nil
}

// get clients' share records of an experiment
func (storage *Storage) GetAllClients(exp_id string) ([]Client, error) {
	var clients []Client
	r := storage.db.Find(&clients, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return clients, nil
}

// get clients' registration records of an experiment
func (storage *Storage) GetRegisteredClient(exp_id string) ([]ClientRegistry, error) {
	var registrations []ClientRegistry
	r := storage.db.Find(&registrations, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}
	return registrations, nil
}

// get all experiments records
func (storage *Storage) GetAllExps() ([]Experiment, error) {
	var experiments []Experiment
	r := storage.db.Find(&experiments, "completed=?", false)
	if r.Error != nil {
		return nil, r.Error
	}

	return experiments, nil
}

// get servers' sumShare records of an experiment
func (storage *Storage) GetAllServers(exp_id string) ([]Server, error) {
	var servers []Server
	r := storage.db.Find(&servers, "exp_id = ?", exp_id)
	if r.Error != nil {
		return nil, r.Error
	}

	return servers, nil
}

// set experiment status to completed
func (storage *Storage) UpdateCompletedExperiment(exp_id string) error {
	var exp Experiment
	r := storage.db.Model(&exp).Where("exp_ID = ?", exp_id).Update("Completed", true)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete experiment record from experiment table
func (storage *Storage) DeleteExperiment(exp_id string) error {
	r := storage.db.Delete(&Experiment{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete client record from client table
func (storage *Storage) DeleteClient(exp_id string) error {
	r := storage.db.Delete(&Client{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete server record from server table
func (storage *Storage) DeleteServer(exp_id string) error {
	r := storage.db.Delete(&Server{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}

// delete client registration record from client table
func (storage *Storage) DeleteClientRegistry(exp_id string) error {
	r := storage.db.Delete(&ClientRegistry{Exp_ID: exp_id})
	if r.Error != nil {
		return r.Error
	}
	return nil
}
