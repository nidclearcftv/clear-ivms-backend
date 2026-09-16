package model

// IVMSType identifies which IVMS vendor a resource originates from, e.g.
// the encoding scheme adapter/cmsv6 uses for its own IDs
// (see adapter/cmsv6/type_id.go).
type IVMSType int

const (
	IVMSTypeUnknown IVMSType = iota
	IVMSTypeCMSV6
)
