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
	Shares    []byte `gorm:"type:longblob"`
}
