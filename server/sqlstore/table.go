package sqlstore

type ClientRegistry struct {
	Exp_ID    string `gorm:"primaryKey"`
	Client_ID string `gorm:"primaryKey"`
	Token     string
}

type Client struct {
	Exp_ID    string `gorm:"primaryKey"`
	Client_ID string `gorm:"primaryKey"`
}

type ClientShare struct {
	Exp_ID    string `gorm:"primaryKey"`
	Client_ID string `gorm:"primaryKey"`
	Shares    []byte `gorm:"type:longblob"`
}

type Complaint struct {
	Exp_ID    string `gorm:"primaryKey"`
	Server_ID string `gorm:"primaryKey"`
	Client_ID string `gorm:"primaryKey"`
	Root      []byte
	Complain  bool
}

type ValidClient struct {
	Exp_ID    string `gorm:"primaryKey"`
	Client_ID string `gorm:"primaryKey"`
}

/**
type Mask struct {
	Exp_ID      string `gorm:"primaryKey"`
	Client_ID   string `gorm:"primaryKey"`
	Input_Index int    `gorm:"primaryKey"`
	Index       int    `gorm:"primaryKey"`
	Value       int
}**/

type MaskedShare struct {
	Exp_ID    string `gorm:"primaryKey"`
	Server_ID string `gorm:"primaryKey"`
	Client_ID string `gorm:"primaryKey"`
	Shares    []byte `gorm:"type:longblob"`
}

type Experiment struct {
	Exp_ID            string `gorm:"primaryKey"`
	ClientShareDue    string
	ComplaintDue      string
	ShareBroadcastDue string
	Owner             string
	Round1_Completed  bool //round1: client share submission
	Round2_Completed  bool //round2:complaint broadcast
	Round3_Completed  bool //round3:masked shares broadcast
}

type EchoComplaint struct {
	Exp_ID     string `gorm:"primaryKey"`
	Server_ID  string `gorm:"primaryKey"`
	Complaints string `gorm:"primaryKey"`
}

type EchoMaskedShare struct {
	Exp_ID       string `gorm:"primaryKey"`
	Server_ID    string `gorm:"primaryKey"`
	MaskedShares string `gorm:"primaryKey"`
}
