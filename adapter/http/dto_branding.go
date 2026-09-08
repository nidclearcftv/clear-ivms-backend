package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// BrandingDTO is the wire representation of model.Branding.
type BrandingDTO struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Domain    string     `json:"domain"`
	Config    model.JSON `json:"config"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func newBrandingDTO(b model.Branding) BrandingDTO {
	return BrandingDTO{
		ID:        string(b.ID),
		Name:      b.Name,
		Domain:    b.Domain,
		Config:    b.Config,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func newBrandingListDTO(list model.List[model.Branding]) ListDTO[BrandingDTO] {
	items := make([]BrandingDTO, len(list.Items))
	for i, b := range list.Items {
		items[i] = newBrandingDTO(b)
	}
	return ListDTO[BrandingDTO]{Items: items, Total: list.Total}
}
