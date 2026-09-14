package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

// registerAccountRoutes registers the admin-only GET /accounts listing —
// e.g. used to search for an account to add to an organization or group.
// requireRolesMiddleware is called with no allowed roles, so only
// model.AccountTypeAdmin — which always bypasses the check — can reach it.
func registerAccountRoutes(rg *gin.RouterGroup, accounts port.AccountService) {
	g := rg.Group("/accounts", authMiddleware(accounts), requireRolesMiddleware())

	g.GET("", func(c *gin.Context) {
		var filters model.AccountFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := accounts.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountListDTO(list))
	})
}
