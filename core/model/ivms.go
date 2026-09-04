package model

type IVMSType int

const (
	IVMSTypeUnknown IVMSType = iota
	IVMSTypeCMSV6
)

// ivmsTypeNames mirrors the vehicles.ivms_type CHECK constraint in the
// Postgres schema.
var ivmsTypeNames = map[IVMSType]string{
	IVMSTypeCMSV6: "cmsv6",
}

func (t IVMSType) String() string {
	return ivmsTypeNames[t]
}

// IVMSTypeFromString parses the Postgres vehicles.ivms_type column back
// into an IVMSType, returning IVMSTypeUnknown for anything unrecognized.
func IVMSTypeFromString(s string) IVMSType {
	for t, name := range ivmsTypeNames {
		if name == s {
			return t
		}
	}
	return IVMSTypeUnknown
}
