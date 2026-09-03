package httpapi

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
)

// ginContextAccountKey is the gin.Context key authMiddleware stores the
// authenticated account under, for handlers to read back via
// contextAccount without hitting the database again.
const ginContextAccountKey = "auth.account"

// orgIDParam is the route param name authMiddleware looks for to scope a
// request to an organization, e.g. a route registered as
// "/organizations/:orgId/...".
const orgIDParam = "orgId"

// authMiddleware validates the session cookie and, on success, makes the
// authenticated account available to the rest of the request: its ID goes
// onto the request's context (see utils.AccountID, read by anything the
// request's context reaches — services, repositories, ...), and the full
// model.Account is stored on the gin.Context (see contextAccount) so
// handlers in this package can use it without a second lookup.
//
// If the matched route has an ":orgId" param, the account must either be
// an admin (model.AccountTypeAdmin, which bypasses the check — an
// org_admin does not) or belong to that organization; otherwise the
// request is rejected with ErrCodeForbidden. On success the organization
// ID is added to the context too (see utils.OrganizationID), same as
// AccountID — this is what lets TeamService/FleetService.List scope to it
// automatically. Routes with no ":orgId" param skip this check entirely.
//
// On any failure, it aborts the request with the mapped error status;
// downstream handlers never run.
//
// Not applied to any route by default — register it explicitly per route
// or route group.
func authMiddleware(accounts port.AccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookieName)
		if err != nil {
			Fail(c, model.ErrCodeInvalidCredentials)
			c.Abort()
			return
		}

		account, err := accounts.Authenticate(c.Request.Context(), token)
		if err != nil {
			RespondError(c, err)
			c.Abort()
			return
		}

		ctx := utils.WithAccountID(c.Request.Context(), account.ID)

		if orgID := model.ID(c.Param(orgIDParam)); orgID != "" {
			if account.Type != model.AccountTypeAdmin {
				member, err := accountBelongsToOrganization(ctx, accounts, account.ID, orgID)
				if err != nil {
					RespondError(c, err)
					c.Abort()
					return
				}
				if !member {
					Fail(c, model.ErrCodeForbidden)
					c.Abort()
					return
				}
			}

			ctx = utils.WithOrganizationID(ctx, orgID)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Set(ginContextAccountKey, account)

		c.Next()
	}
}

// accountBelongsToOrganization reports whether accountID is a member of
// organizationID. There's no dedicated membership-check port method, so
// this lists the organization's accounts and scans for a match — fine at
// today's scale, worth a direct "IsMember" repository method if
// organizations grow large enough for this to matter.
func accountBelongsToOrganization(ctx context.Context, accounts port.AccountService, accountID, organizationID model.ID) (bool, error) {
	members, err := accounts.ListFromOrganization(ctx, organizationID)
	if err != nil {
		return false, err
	}

	for _, member := range members.Items {
		if member.ID == accountID {
			return true, nil
		}
	}

	return false, nil
}

// contextAccount returns the account authMiddleware attached to c.
func contextAccount(c *gin.Context) (model.Account, bool) {
	v, ok := c.Get(ginContextAccountKey)
	if !ok {
		return model.Account{}, false
	}
	account, ok := v.(model.Account)
	return account, ok
}
