package sqlstore

type Experiment struct {
	Exp_ID         string `gorm:"primaryKey"`
	ClientShareDue string
	ServerShareDue string
	Completed      bool
}

type ServerShare struct {
	Exp_ID    string `gorm:"primaryKey"`
	Server_ID string `gorm:"primaryKey"`
	Index     int    `gorm:"primaryKey"`
	Value     int
}

/**
type ServerComputation struct {
	Exp_ID         string
	Server_ID      string
	SumShare_Value int
	SumShare_Index int
}**/
