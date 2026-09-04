package httpapi

import "github.com/nidclearcftv/clear-ivms-backend/core/model"

// ListDTO is the wire representation of a paginated, model.List[T]-shaped
// result.
type ListDTO[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

// nullableIDFromRequest returns nil for an empty id — no value — and a
// pointer to it otherwise. Used for optional *model.ID fields bound from
// JSON request bodies (e.g. a group's parentId, a vehicle's groupId).
func nullableIDFromRequest(id string) *model.ID {
	if id == "" {
		return nil
	}
	v := model.ID(id)
	return &v
}
