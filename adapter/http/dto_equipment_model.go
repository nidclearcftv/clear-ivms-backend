package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// EquipmentModelDTO is the wire representation of model.EquipmentModel.
type EquipmentModelDTO struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Type           string    `json:"type"`
	Public         bool      `json:"public"`
	OrganizationID string    `json:"organizationId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func newEquipmentModelDTO(m model.EquipmentModel) EquipmentModelDTO {
	return EquipmentModelDTO{
		ID:             string(m.ID),
		Name:           m.Name,
		Description:    m.Description,
		Type:           string(m.Type),
		Public:         m.Public,
		OrganizationID: string(m.OrganizationID),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func newEquipmentModelListDTO(list model.List[model.EquipmentModel]) ListDTO[EquipmentModelDTO] {
	items := make([]EquipmentModelDTO, len(list.Items))
	for i, m := range list.Items {
		items[i] = newEquipmentModelDTO(m)
	}
	return ListDTO[EquipmentModelDTO]{Items: items, Total: list.Total}
}
