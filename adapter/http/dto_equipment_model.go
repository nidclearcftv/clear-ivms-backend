package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// EquipmentModelDTO is the wire representation of model.EquipmentModel.
// HasPicture reports whether GET .../picture will return an image —
// PictureObjectKey itself is a storage implementation detail and never
// leaves the backend.
type EquipmentModelDTO struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Manufacturer    string    `json:"manufacturer"`
	Features        []string  `json:"features"`
	Type            string    `json:"type"`
	Public          bool      `json:"public"`
	HasPicture      bool      `json:"hasPicture"`
	ExternalViewURL string    `json:"externalViewUrl"`
	OrganizationID  string    `json:"organizationId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func newEquipmentModelDTO(m model.EquipmentModel) EquipmentModelDTO {
	features := m.Features
	if features == nil {
		features = []string{}
	}
	return EquipmentModelDTO{
		ID:              string(m.ID),
		Name:            m.Name,
		Description:     m.Description,
		Manufacturer:    m.Manufacturer,
		Features:        features,
		Type:            string(m.Type),
		Public:          m.Public,
		HasPicture:      m.PictureObjectKey != nil,
		ExternalViewURL: m.ExternalViewURL,
		OrganizationID:  string(m.OrganizationID),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func newEquipmentModelListDTO(list model.List[model.EquipmentModel]) ListDTO[EquipmentModelDTO] {
	items := make([]EquipmentModelDTO, len(list.Items))
	for i, m := range list.Items {
		items[i] = newEquipmentModelDTO(m)
	}
	return ListDTO[EquipmentModelDTO]{Items: items, Total: list.Total}
}
