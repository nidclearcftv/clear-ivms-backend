package model

type IVMSType int

const (
	IVMSTypeUnknown IVMSType = iota
	IVMSTypeCMSV6
	// IVMSTypeNone means the vehicle isn't tracked by any IVMS vendor
	// integration (e.g. added manually) — a legitimate stored value,
	// unlike IVMSTypeUnknown, which is only ever a parse-failure fallback.
	IVMSTypeNone
)

// ivmsTypeNames mirrors the vehicles.ivms_type CHECK constraint in the
// Postgres schema.
var ivmsTypeNames = map[IVMSType]string{
	IVMSTypeCMSV6: "cmsv6",
	IVMSTypeNone:  "none",
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
