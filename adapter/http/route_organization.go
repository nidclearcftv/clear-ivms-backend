package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

type createOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
}

// registerOrganizationRoutes registers /organizations under rg, restricted
// to admins: requireRolesMiddleware is called with no allowed roles, so
// only model.AccountTypeAdmin — which always bypasses the check — can reach
// any of them.
func registerOrganizationRoutes(rg *gin.RouterGroup, organizations port.OrganizationService, accounts port.AccountService) {
	g := rg.Group("/organizations", authMiddleware(accounts), requireRolesMiddleware())

	g.POST("", func(c *gin.Context) {
		var req createOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		organization, err := organizations.Create(c.Request.Context(), model.Organization{Name: req.Name})
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newOrganizationDTO(organization))
	})

	g.GET("", func(c *gin.Context) {
		var filters model.OrganizationFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := organizations.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newOrganizationListDTO(list))
	})

	g.GET("/:id", func(c *gin.Context) {
		organization, err := organizations.Get(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newOrganizationDTO(organization))
	})

	g.PUT("/:id", func(c *gin.Context) {
		var req updateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		organization, err := organizations.Update(c.Request.Context(), model.Organization{
			ID:   model.ID(c.Param("id")),
			Name: req.Name,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newOrganizationDTO(organization))
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := organizations.Delete(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.GET("/:id/accounts", func(c *gin.Context) {
		list, err := accounts.ListFromOrganization(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountListDTO(list))
	})

	g.POST("/:id/accounts/:accountId", func(c *gin.Context) {
		err := organizations.AddAccount(c.Request.Context(), model.ID(c.Param("id")), model.ID(c.Param("accountId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.DELETE("/:id/accounts/:accountId", func(c *gin.Context) {
		err := organizations.RemoveAccount(c.Request.Context(), model.ID(c.Param("id")), model.ID(c.Param("accountId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})
}
