package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
)

type createGroupRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID string `json:"parentId"`
}

type updateGroupRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID string `json:"parentId"`
}

// registerGroupRoutes registers /groups under rg, gated by requireOrganizationMiddleware
// (mandatory "?orgId="; see that middleware for the membership check it does)
// and requireRolesMiddleware(model.AccountTypeOrgAdmin) — so besides
// model.AccountTypeAdmin, which always bypasses role checks, an org_admin
// of the request's organization can reach these routes too; a plain user
// cannot.
func registerGroupRoutes(rg *gin.RouterGroup, groups port.GroupService, accounts port.AccountService) {
	g := rg.Group("/groups",
		authMiddleware(accounts),
		requireRolesMiddleware(model.AccountTypeOrgAdmin),
		requireOrganizationMiddleware(accounts),
	)

	g.POST("", func(c *gin.Context) {
		var req createGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		group, err := groups.Create(c.Request.Context(), model.Group{
			Name:           req.Name,
			OrganizationID: utils.OrganizationID(c.Request.Context()),
			ParentID:       nullableIDFromRequest(req.ParentID),
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newGroupDTO(group))
	})

	g.GET("", func(c *gin.Context) {
		var filters model.GroupFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := groups.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newGroupListDTO(list))
	})

	g.GET("/:id", func(c *gin.Context) {
		group, err := groups.Get(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newGroupDTO(group))
	})

	g.PUT("/:id", func(c *gin.Context) {
		var req updateGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		group, err := groups.Update(c.Request.Context(), model.Group{
			ID:             model.ID(c.Param("id")),
			Name:           req.Name,
			OrganizationID: utils.OrganizationID(c.Request.Context()),
			ParentID:       nullableIDFromRequest(req.ParentID),
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newGroupDTO(group))
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := groups.Delete(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})
}
