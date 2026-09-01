package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleDTO is the wire representation of model.Vehicle. It intentionally
// excludes internal-only data — model.Vehicle.Attributes (the raw,
// source-specific payload) and IVMSType (which internal adapter sourced
// the record) — so vendor/adapter details never leak to API consumers.
type VehicleDTO struct {
	ID          string    `json:"id"`
	FleetID     string    `json:"fleetId"`
	Type        string    `json:"type"`
	PlateNumber string    `json:"plateNumber"`
	DeviceTime  time.Time `json:"deviceTime"`
}

func newVehicleDTO(v model.Vehicle) VehicleDTO {
	return VehicleDTO{
		ID:          string(v.ID),
		FleetID:     string(v.FleetID),
		Type:        v.Type,
		PlateNumber: v.PlateNumber,
		DeviceTime:  v.DeviceTime,
	}
}

func newVehicleListDTO(list model.List[model.Vehicle]) ListDTO[VehicleDTO] {
	items := make([]VehicleDTO, len(list.Items))
	for i, v := range list.Items {
		items[i] = newVehicleDTO(v)
	}
	return ListDTO[VehicleDTO]{Items: items, Total: list.Total}
}
