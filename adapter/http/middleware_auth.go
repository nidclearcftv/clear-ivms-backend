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

// orgIDQueryKey is the query string key requireOrganizationMiddleware
// looks for to scope a request to an organization, e.g. "?orgId=...".
const orgIDQueryKey = "orgId"

// authMiddleware validates the session cookie and, on success, makes the
// authenticated account available to the rest of the request: its ID goes
// onto the request's context (see utils.AccountID, read by anything the
// request's context reaches — services, repositories, ...), and the full
// model.Account is stored on the gin.Context (see contextAccount) so
// handlers in this package — including requireOrganizationMiddleware
// below — can use it without a second lookup.
//
// On failure, it aborts the request with the mapped error status;
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

		c.Request = c.Request.WithContext(utils.WithAccountID(c.Request.Context(), account.ID))
		c.Set(ginContextAccountKey, account)

		c.Next()
	}
}

// requireOrganizationMiddleware must run after authMiddleware — it reads
// the account authMiddleware attaches via contextAccount, and rejects the
// request with ErrCodeUnknown if that's missing (a wiring error: the two
// were not chained correctly).
//
// It requires an "?orgId=" query param, rejecting the request with
// ErrCodeInvalidRequest if it's absent — unlike authMiddleware's old
// inline version of this check, organization scoping is mandatory on any
// route this is applied to. Unless the account is an admin
// (model.AccountTypeAdmin — org_admin does not bypass this), the account
// must belong to that organization or the request is rejected with
// ErrCodeForbidden. On success the organization ID is added to the
// context (see utils.OrganizationID), which is what lets
// TeamService/FleetService.List scope to it automatically.
//
// Not applied to any route by default — register it explicitly, after
// authMiddleware, per route or route group.
func requireOrganizationMiddleware(accounts port.AccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := model.ID(c.Query(orgIDQueryKey))
		if orgID == "" {
			Fail(c, model.ErrCodeInvalidRequest, "missing required query parameter: "+orgIDQueryKey)
			c.Abort()
			return
		}

		account, ok := contextAccount(c)
		if !ok {
			Fail(c, model.ErrCodeUnknown)
			c.Abort()
			return
		}

		if account.Type != model.AccountTypeAdmin {
			member, err := accounts.IsMemberOfOrganization(c.Request.Context(), account.ID, orgID)
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

		c.Request = c.Request.WithContext(utils.WithOrganizationID(c.Request.Context(), orgID))
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
