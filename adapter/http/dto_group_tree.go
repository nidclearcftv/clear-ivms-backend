package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// GroupTreeVehicleDTO is the wire representation of a vehicle inside a
// GroupTreeDTO — deliberately its own type (same field set as VehicleDTO)
// so this response's shape doesn't silently change if VehicleDTO's does.
type GroupTreeVehicleDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PlateNumber string    `json:"plateNumber"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// GroupTreeNodeDTO is one node of GroupTreeDTO. ParentID/OrganizationID are
// deliberately omitted — implied by tree position, never needed on the
// wire.
type GroupTreeNodeDTO struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	Children  []GroupTreeNodeDTO    `json:"children"`
	Vehicles  []GroupTreeVehicleDTO `json:"vehicles"`
}

// GroupTreeDTO is the wire representation of model.GroupTree.
type GroupTreeDTO struct {
	Roots              []GroupTreeNodeDTO    `json:"roots"`
	UnassignedVehicles []GroupTreeVehicleDTO `json:"unassignedVehicles"`
}

func newGroupTreeVehicleDTO(v model.Vehicle) GroupTreeVehicleDTO {
	return GroupTreeVehicleDTO{
		ID:          string(v.ID),
		Name:        v.Name,
		PlateNumber: v.PlateNumber,
		Status:      string(v.Status),
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func newGroupTreeNodeDTO(n model.GroupTreeNode) GroupTreeNodeDTO {
	children := make([]GroupTreeNodeDTO, len(n.Children))
	for i, c := range n.Children {
		children[i] = newGroupTreeNodeDTO(c)
	}
	vehicles := make([]GroupTreeVehicleDTO, len(n.Vehicles))
	for i, v := range n.Vehicles {
		vehicles[i] = newGroupTreeVehicleDTO(v)
	}
	return GroupTreeNodeDTO{
		ID:        string(n.Group.ID),
		Name:      n.Group.Name,
		CreatedAt: n.Group.CreatedAt,
		UpdatedAt: n.Group.UpdatedAt,
		Children:  children,
		Vehicles:  vehicles,
	}
}

func newGroupTreeDTO(t model.GroupTree) GroupTreeDTO {
	roots := make([]GroupTreeNodeDTO, len(t.Roots))
	for i, r := range t.Roots {
		roots[i] = newGroupTreeNodeDTO(r)
	}
	unassigned := make([]GroupTreeVehicleDTO, len(t.UnassignedVehicles))
	for i, v := range t.UnassignedVehicles {
		unassigned[i] = newGroupTreeVehicleDTO(v)
	}
	return GroupTreeDTO{Roots: roots, UnassignedVehicles: unassigned}
}
