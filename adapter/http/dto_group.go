package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// GroupDTO is the wire representation of model.Group.
type GroupDTO struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organizationId"`
	ParentID       string    `json:"parentId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func newGroupDTO(g model.Group) GroupDTO {
	dto := GroupDTO{
		ID:             string(g.ID),
		Name:           g.Name,
		OrganizationID: string(g.OrganizationID),
		CreatedAt:      g.CreatedAt,
		UpdatedAt:      g.UpdatedAt,
	}
	if g.ParentID != nil {
		dto.ParentID = string(*g.ParentID)
	}
	return dto
}

func newGroupListDTO(list model.List[model.Group]) ListDTO[GroupDTO] {
	items := make([]GroupDTO, len(list.Items))
	for i, g := range list.Items {
		items[i] = newGroupDTO(g)
	}
	return ListDTO[GroupDTO]{Items: items, Total: list.Total}
}
