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
