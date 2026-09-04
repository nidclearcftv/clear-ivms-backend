package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleDTO is the wire representation of model.Vehicle. It intentionally
// excludes IVMSType and ExternalID — which vendor sourced the record, and
// its ID within that vendor's system — so vendor/adapter details never leak
// to API consumers.
type VehicleDTO struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	FleetID        string    `json:"fleetId,omitempty"`
	PlateNumber    string    `json:"plateNumber"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func newVehicleDTO(v model.Vehicle) VehicleDTO {
	dto := VehicleDTO{
		ID:             string(v.ID),
		OrganizationID: string(v.OrganizationID),
		PlateNumber:    v.PlateNumber,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
	if v.FleetID != nil {
		dto.FleetID = string(*v.FleetID)
	}
	return dto
}

func newVehicleListDTO(list model.List[model.Vehicle]) ListDTO[VehicleDTO] {
	items := make([]VehicleDTO, len(list.Items))
	for i, v := range list.Items {
		items[i] = newVehicleDTO(v)
	}
	return ListDTO[VehicleDTO]{Items: items, Total: list.Total}
}
