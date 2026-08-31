package cmsv6

// LoginResponse is the payload returned by /login/loginMobile.do.
type LoginResponse struct {
	H5ClientSSLEnabled int       `json:"h5ClientSSLEnabled"`
	AddrPort           int       `json:"addrPort"`
	WsPortSSL          int       `json:"wsPortSSL"`
	IsTtsPlan          int       `json:"isTtsPlan"`
	ServerVersion      string    `json:"serverVersion"`
	Useid              int       `json:"useid"`
	Session            string    `json:"session"`
	AddrServerIp       string    `json:"addrServerIp"`
	ServerPort         int       `json:"serverPort"`
	H5IPClientSSL      string    `json:"h5IPClientSSL"`
	Privis             []int     `json:"privis"`
	Result             int       `json:"result"`
	WsIp               string    `json:"wsIp"`
	RemindDay          int       `json:"remindDay"`
	TalkServerPort     int       `json:"talkServerPort"`
	AddrServerPort     string    `json:"addrServerPort"`
	TalkServerIp       string    `json:"talkServerIp"`
	VerFile            string    `json:"verFile"`
	H5Ip               string    `json:"h5Ip"`
	H5Port             int       `json:"h5Port"`
	ServerLanIp        string    `json:"serverLanIp"`
	VerName            string    `json:"verName"`
	H5PortClientSSL    int       `json:"h5PortClientSSL"`
	IsAdmin            int       `json:"isAdmin"`
	H5PortClient       int       `json:"h5PortClient"`
	RemindMessage      string    `json:"remindMessage"`
	Url                string    `json:"url"`
	VerCode            string    `json:"verCode"`
	Pojo               LoginPojo `json:"pojo"`
	SplitAlarmTable    int       `json:"splitAlarmTable"`
	PwdStatus          int       `json:"pwdStatus"`
	WsPort             int       `json:"wsPort"`
	Name               string    `json:"name"`
	ServerIp           string    `json:"serverIp"`
	H5PortSSL          int       `json:"h5PortSSL"`
	IsTurkeyTSG        int       `json:"isTurkeyTSG"`
	Account            string    `json:"account"`
	H5IPClient         string    `json:"h5IPClient"`
}

type LoginPojo struct {
	ID              int         `json:"id"`
	Aid             int         `json:"aid"`
	UserId          int         `json:"userId"`
	Account         string      `json:"account"`
	Name            string      `json:"name"`
	IsAdmin         int         `json:"isAdmin"`
	Ip              string      `json:"ip"`
	Pid             int         `json:"pid"`
	Rid             string      `json:"rid"`
	Privilege       interface{} `json:"privilege"`
	Url             string      `json:"url"`
	PwdStatus       int         `json:"pwdStatus"`
	LastLoginTime   int64       `json:"lastLoginTime"`
	Validity        int64       `json:"validity"`
	SessionId       string      `json:"sessionId"`
	Level           int         `json:"level"`
	Location        interface{} `json:"location"`
	Zoom            interface{} `json:"zoom"`
	ValidTime       interface{} `json:"validTime"`
	SpeedUnits      interface{} `json:"speedUnits"`
	MultiPointLogin interface{} `json:"multiPointLogin"`
	Authorization   interface{} `json:"authorization"`
}
