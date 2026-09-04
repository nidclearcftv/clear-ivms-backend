package utils

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// ctxKey is an unexported type for context keys defined in this package,
// so they can never collide with keys set by other packages (the risk
// with using a raw string, e.g. ctx.Value("agent_id"), directly).
type ctxKey int

const (
	ctxKeyAccountID ctxKey = iota
	ctxKeyOrganizationID
)

// WithAccountID returns a copy of ctx carrying the authenticated account's
// ID, for AccountID to read back later in the request's lifecycle.
func WithAccountID(ctx context.Context, id model.ID) context.Context {
	return context.WithValue(ctx, ctxKeyAccountID, id)
}

// AccountID returns the authenticated account's ID carried on ctx, or ""
// if none was set.
func AccountID(ctx context.Context) model.ID {
	id, _ := ctx.Value(ctxKeyAccountID).(model.ID)
	return id
}

// WithOrganizationID returns a copy of ctx carrying the current request's
// organization ID, for OrganizationID to read back later. Services use
// this to scope organization-owned resources (e.g. groups) automatically,
// rather than trusting a caller-supplied organization ID.
func WithOrganizationID(ctx context.Context, id model.ID) context.Context {
	return context.WithValue(ctx, ctxKeyOrganizationID, id)
}

// OrganizationID returns the current organization ID carried on ctx, or ""
// if none was set.
func OrganizationID(ctx context.Context) model.ID {
	id, _ := ctx.Value(ctxKeyOrganizationID).(model.ID)
	return id
}
