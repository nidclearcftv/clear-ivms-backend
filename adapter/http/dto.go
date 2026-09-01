package httpapi

// ListDTO is the wire representation of a paginated, model.List[T]-shaped
// result.
type ListDTO[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}
