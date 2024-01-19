package sqlstore

type Client struct {
	Exp_ID    string
	Client_ID string
}

type ClientShare struct {
	Exp_ID      string
	Client_ID   string
	Input_Index int
	Share_Index int
	Share_Value int
}

type Complaint struct {
	Exp_ID    string
	Server_ID string
	Client_ID string
	Complain  bool
	Root      []byte
}

type ValidClient struct {
	Exp_ID    string
	Client_ID string
}

type Mask struct {
	Exp_ID      string
	Client_ID   string
	Input_Index int
	Index       int
	Value       int
}

type MaskedShare struct {
	Exp_ID      string
	Server_ID   string
	Client_ID   string
	Input_Index int
	Index       int
	Value       int
}

type Experiment struct {
	Exp_ID            string
	ClientShareDue    string
	ComplaintDue      string
	ShareBroadcastDue string
	Owner             string
	Round1_Completed  bool //round1: client share submission
	Round2_Completed  bool //round2:complaint broadcast
	Round3_Completed  bool //round3:masked shares broadcast
}

/**

type ServerComputation struct {
	Exp_ID         string
	Server_ID      string
	SumShare_Value int
	SumShare_Index int
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

**/
