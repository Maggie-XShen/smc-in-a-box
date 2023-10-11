package sqlstore

import "gorm.io/datatypes"

type ClientShare struct {
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

type ClientSet struct {
	Exp_ID    string
	Server_ID string
	Clients   datatypes.JSON
}

type Experiment struct {
	Exp_ID                 string
	Due                    string
	Owner                  string
	Server_Round_Completed bool
	Completed              bool
}

type ServerComputation struct {
	Exp_ID         string
	Server_ID      string
	SumShare_Value int
	SumShare_Index int
}
