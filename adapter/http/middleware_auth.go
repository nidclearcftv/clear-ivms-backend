package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
)

// ginContextAccountKey is the gin.Context key authMiddleware stores the
// authenticated account under, for handlers to read back via
// contextAccount without hitting the database again.
const ginContextAccountKey = "auth.account"

// orgIDQueryKey is the query string key authMiddleware looks for to scope
// a request to an organization, e.g. "?orgId=...".
const orgIDQueryKey = "orgId"

// authMiddleware validates the session cookie and, on success, makes the
// authenticated account available to the rest of the request: its ID goes
// onto the request's context (see utils.AccountID, read by anything the
// request's context reaches — services, repositories, ...), and the full
// model.Account is stored on the gin.Context (see contextAccount) so
// handlers in this package can use it without a second lookup.
//
// If the request has an "?orgId=" query param, the account must either be
// an admin (model.AccountTypeAdmin, which bypasses the check — an
// org_admin does not) or belong to that organization; otherwise the
// request is rejected with ErrCodeForbidden. On success the organization
// ID is added to the context too (see utils.OrganizationID), same as
// AccountID — this is what lets TeamService/FleetService.List scope to it
// automatically. Requests with no "orgId" query param skip this check
// entirely.
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

		if orgID := model.ID(c.Query(orgIDQueryKey)); orgID != "" {
			if account.Type != model.AccountTypeAdmin {
				member, err := accounts.IsMemberOfOrganization(ctx, account.ID, orgID)
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

// contextAccount returns the account authMiddleware attached to c.
func contextAccount(c *gin.Context) (model.Account, bool) {
	v, ok := c.Get(ginContextAccountKey)
	if !ok {
		return model.Account{}, false
	}
	account, ok := v.(model.Account)
	return account, ok
}
