package cmsv6

// VehicleListResponse is the payload returned by /vehicle/list.do.
type VehicleListResponse struct {
	UserDeviceCount int       `json:"userDeviceCount"`
	UserDeviceList  []Vehicle `json:"userDeviceList"`
}

// Vehicle is a single entry of VehicleListResponse.UserDeviceList.
//
// A number of fields are only ever observed as null in sample payloads, so
// their real type is unknown; those are kept as interface{} rather than
// guessing a concrete type that could fail to unmarshal once populated.
type Vehicle struct {
	ID          int    `json:"ID"`
	IDNO        string `json:"IDNO"`
	DevID       string `json:"DevID"`
	DevType     int    `json:"DevType"`
	DevStatus   int    `json:"DevStatus"`
	DevGroupID  int    `json:"DevGroupID"`
	Protocol    int    `json:"Protocol"`
	Module      int    `json:"Module"`
	IMEI        string `json:"IMEI"`
	SimCard     string `json:"SimCard"`
	NetAddrType int    `json:"NetAddrType"`
	StopUsing   int    `json:"stopUsing"`

	PlateIDNO   string `json:"PlateIDNO"`
	PlateColor  int    `json:"PlateColor"`
	VehiType    string `json:"VehiType"`
	VehiColor   string `json:"VehiColor"`
	VehiCompany string `json:"VehiCompany"`

	ChnCount int    `json:"ChnCount"`
	ChnName  string `json:"ChnName"`

	IOInCount     int    `json:"IOInCount"`
	IOInName      string `json:"IOInName"`
	IoInName      string `json:"ioInName"`
	TempCount     int    `json:"TempCount"`
	TempName      string `json:"TempName"`
	HumidityCount int    `json:"HumidityCount"`
	HumidityName  string `json:"HumidityName"`

	EnabledChannels  int64 `json:"EnabledChannels"`
	PrivacyChannels  int   `json:"privacyChannels"`
	PeripheralStatus int   `json:"peripheral"`

	AccountID    int    `json:"AccountID"`
	UserID       int    `json:"UserID"`
	ParentID     int    `json:"ParentID"`
	GroupId      int    `json:"groupId"`
	GroupName    string `json:"groupName"`
	DevGroupName string `json:"devGroupName"`

	DriverId   string `json:"DriverId"`
	DriverName string `json:"DriverName"`
	DriverTele string `json:"DriverTele"`

	Icon            int     `json:"Icon"`
	TerminalModel   string  `json:"TerminalModel"`
	DateProduct     int64   `json:"DateProduct"`
	Capacity        float64 `json:"capacity"`
	OilCapacity     float64 `json:"oilCapacity"`
	RatedLoad       float64 `json:"RatedLoad"`
	RatedPassenger  int     `json:"RatedPassenger"`
	NationalityCode int     `json:"NationalityCode"`
	PermitNo        int     `json:"permitNo"`
	IsIvmsSupported int     `json:"isIvmsSupported"`
	SimCarMode      int     `json:"simCarMode"`
	Bind            int     `json:"bind"`

	Address          string `json:"address"`
	Region           string `json:"region"`
	Telephone        string `json:"telephone"`
	Linkman          string `json:"linkman"`
	LegalPerson      string `json:"legalPerson"`
	LegalPhone       string `json:"legalPhone"`
	BusinessScope    string `json:"businessScope"`
	RoadPermit       string `json:"roadPermit"`
	LogisticsLockSN  string `json:"logisticsLockSN"`
	NuclearAuthority string `json:"nuclearAuthority"`

	JT808Lines              string      `json:"JT808_Lines"`
	JT808VehicleCity        string      `json:"JT808_VehicleCity"`
	JT808TransType          interface{} `json:"JT808_TransType"`
	JT808VehicleNationality interface{} `json:"JT808_VehicleNationality"`

	PayEnable   int         `json:"PayEnable"`
	PayMonth    interface{} `json:"PayMonth"`
	PayPeriod   interface{} `json:"PayPeriod"`
	PayBegin    interface{} `json:"PayBegin"`
	PayDelayDay interface{} `json:"PayDelayDay"`

	DrivingRecorderStandard int `json:"drivingrecorderstandard"`

	IVMSDevIDNO              interface{} `json:"IVMSDevIDNO"`
	DantonCentralProductID   interface{} `json:"danton_centralProductID"`
	DantonCentralEquipmentNo interface{} `json:"danton_centralEquipmentNo"`
	DantonSerialNo           interface{} `json:"danton_serialNo"`
	DantonVehiTypeName       interface{} `json:"danton_vehiTypeName"`
	DantonOrderId            interface{} `json:"danton_orderId"`

	CompanyName          interface{} `json:"companyName"`
	TypeId               interface{} `json:"TypeId"`
	TypeName             interface{} `json:"TypeName"`
	BandName             interface{} `json:"BandName"`
	BandId               interface{} `json:"BandId"`
	Nationality          interface{} `json:"nationality"`
	VehiRegisterNum      interface{} `json:"VehiRegisterNum"`
	InternatLicense      interface{} `json:"InternatLicense"`
	Industry             interface{} `json:"industry"`
	AnnualReview         interface{} `json:"AnnualReview"`
	Insurance            interface{} `json:"insurance"`
	DueDate              interface{} `json:"dueDate"`
	EncrytedChipNo       interface{} `json:"encrytedChipNo"`
	StoDay               interface{} `json:"StoDay"`
	Manufacturer         interface{} `json:"Manufacturer"`
	SuperId              interface{} `json:"superId"`
	Qualification        interface{} `json:"qualification"`
	VehiIDCode           interface{} `json:"VehiIDCode"`
	Remark               interface{} `json:"remark"`
	InstallTime          interface{} `json:"installTime"`
	InstallAddress       interface{} `json:"installAddress"`
	VehiMarking          interface{} `json:"VehiMarking"`
	BDMobileTerminalCode interface{} `json:"BDMobileTerminalCode"`
	DeviceOrientation    interface{} `json:"deviceOrientation"`
}
