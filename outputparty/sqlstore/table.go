package sqlstore

type Experiment struct {
	Exp_ID                 string
	Due                    string
	Owner                  string
	Server_Round_Completed bool
	Completed              bool
}

type AggregatedShare struct {
	Exp_ID    string
	Server_ID string
	Index     int
	Value     int
}

/**
type ServerComputation struct {
	Exp_ID         string
	Server_ID      string
	SumShare_Value int
	SumShare_Index int
}**/
