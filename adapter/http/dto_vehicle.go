package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleDTO is the wire representation of model.Vehicle.
type VehicleDTO struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	GroupID        string    `json:"groupId,omitempty"`
	Name           string    `json:"name"`
	ExternalID     string    `json:"externalId"`
	PlateNumber    string    `json:"plateNumber"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func newVehicleDTO(v model.Vehicle) VehicleDTO {
	dto := VehicleDTO{
		ID:             string(v.ID),
		OrganizationID: string(v.OrganizationID),
		Name:           v.Name,
		ExternalID:     v.ExternalID,
		PlateNumber:    v.PlateNumber,
		Status:         string(v.Status),
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
	if v.GroupID != nil {
		dto.GroupID = string(*v.GroupID)
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
