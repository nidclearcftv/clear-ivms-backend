package httpapi

import (
	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleDTO is the wire representation of model.Vehicle. It intentionally
// excludes internal-only data — model.Vehicle.Attributes (the raw,
// source-specific payload) and IVMSType (which internal adapter sourced
// the record) — so vendor/adapter details never leak to API consumers.
type VehicleDTO struct {
	ID          string `json:"id"`
	FleetID     string `json:"fleetId"`
	PlateNumber string `json:"plateNumber"`
}

func newVehicleDTO(v model.Vehicle) VehicleDTO {
	return VehicleDTO{
		ID:          string(v.ID),
		FleetID:     string(v.FleetID),
		PlateNumber: v.PlateNumber,
	}
}

func newVehicleListDTO(list model.List[model.Vehicle]) ListDTO[VehicleDTO] {
	items := make([]VehicleDTO, len(list.Items))
	for i, v := range list.Items {
		items[i] = newVehicleDTO(v)
	}
	return ListDTO[VehicleDTO]{Items: items, Total: list.Total}
}
