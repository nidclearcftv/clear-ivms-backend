package model

type Vehicle struct {
	ID          ID
	FleetID     ID
	IVMSType    IVMSType
	PlateNumber string
}

type VehicleFilters struct {
}

func (f *VehicleFilters) String() string {
	return ""
}

func VechicleKey(id ID) string {
	return "vehicle:" + string(id)
}

func VechicleListKey(agentID ID, filters VehicleFilters) string {
	if agentID == "" {
		return ""
	}
	return "vehicle_list:" + string(agentID) + ":" + filters.String()
}
