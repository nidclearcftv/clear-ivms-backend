package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// OrganizationDTO is the wire representation of model.Organization.
type OrganizationDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func newOrganizationDTO(o model.Organization) OrganizationDTO {
	return OrganizationDTO{
		ID:        string(o.ID),
		Name:      o.Name,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func newOrganizationListDTO(list model.List[model.Organization]) ListDTO[OrganizationDTO] {
	items := make([]OrganizationDTO, len(list.Items))
	for i, o := range list.Items {
		items[i] = newOrganizationDTO(o)
	}
	return ListDTO[OrganizationDTO]{Items: items, Total: list.Total}
}
