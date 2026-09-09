package httpapi

import (
	"time"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// AccountDTO is the wire representation of model.Account. Its password
// hash was already excluded at the domain-model level (see model.Account),
// so there's nothing extra to strip here.
type AccountDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phoneNumber"`
	Type        string    `json:"type"`
	Blocked     bool      `json:"blocked"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func newAccountDTO(a model.Account) AccountDTO {
	return AccountDTO{
		ID:          string(a.ID),
		Name:        a.Name,
		Email:       a.Email,
		PhoneNumber: a.PhoneNumber,
		Type:        string(a.Type),
		Blocked:     a.Blocked,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func newAccountListDTO(list model.List[model.Account]) ListDTO[AccountDTO] {
	items := make([]AccountDTO, len(list.Items))
	for i, a := range list.Items {
		items[i] = newAccountDTO(a)
	}
	return ListDTO[AccountDTO]{Items: items, Total: list.Total}
}

// OrganizationSummaryDTO is the minimal (id, name) representation of
// model.Organization used by MeDTO — /me only needs enough to populate an
// organization switcher, not the full OrganizationDTO.
type OrganizationSummaryDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MeDTO is the response shape for GET /me: the account plus the
// organizations it belongs to.
type MeDTO struct {
	AccountDTO
	Organizations []OrganizationSummaryDTO `json:"organizations"`
}

func newMeDTO(a model.Account, organizations []model.Organization) MeDTO {
	summaries := make([]OrganizationSummaryDTO, len(organizations))
	for i, o := range organizations {
		summaries[i] = OrganizationSummaryDTO{ID: string(o.ID), Name: o.Name}
	}
	return MeDTO{AccountDTO: newAccountDTO(a), Organizations: summaries}
}
