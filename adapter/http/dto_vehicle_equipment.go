package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleEquipmentDTO is the wire representation of
// model.VehicleEquipment.
type VehicleEquipmentDTO struct {
	ID                 string    `json:"id"`
	VehicleID          string    `json:"vehicleId"`
	EquipmentModelID   string    `json:"equipmentModelId"`
	EquipmentModelType string    `json:"equipmentModelType"`
	SerialNumber       string    `json:"serialNumber"`
	Description        string    `json:"description"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func newVehicleEquipmentDTO(ve model.VehicleEquipment) VehicleEquipmentDTO {
	return VehicleEquipmentDTO{
		ID:                 string(ve.ID),
		VehicleID:          string(ve.VehicleID),
		EquipmentModelID:   string(ve.EquipmentModelID),
		EquipmentModelType: string(ve.EquipmentModelType),
		SerialNumber:       ve.SerialNumber,
		Description:        ve.Description,
		CreatedAt:          ve.CreatedAt,
		UpdatedAt:          ve.UpdatedAt,
	}
}

// newVehicleEquipmentListDTO wraps items in the same ListDTO envelope
// every other list endpoint uses — Total is just len(items), since this
// list is never paginated (see port.VehicleEquipmentRepository.List).
func newVehicleEquipmentListDTO(items []model.VehicleEquipment) ListDTO[VehicleEquipmentDTO] {
	dtos := make([]VehicleEquipmentDTO, len(items))
	for i, ve := range items {
		dtos[i] = newVehicleEquipmentDTO(ve)
	}
	return ListDTO[VehicleEquipmentDTO]{Items: dtos, Total: len(dtos)}
}
