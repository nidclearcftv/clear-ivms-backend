package model

import "time"

type Vehicle struct {
	ID          ID
	FleetID     ID
	IVMSType    IVMSType
	Type        string
	PlateNumber string
	LastSeen    time.Time
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
