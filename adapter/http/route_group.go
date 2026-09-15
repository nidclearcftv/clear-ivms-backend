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

	// GetTree assembles the whole organization's fleet hierarchy (every
	// group nested under its parent, every vehicle under its group, plus
	// an UnassignedVehicles list) in one call — see
	// port.GroupService.GetTree.
	g.GET("/tree", func(c *gin.Context) {
		tree, err := groups.GetTree(c.Request.Context())
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newGroupTreeDTO(tree))
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

	g.GET("/:id/accounts", func(c *gin.Context) {
		list, err := accounts.ListFromGroup(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountListDTO(list))
	})

	// AddAccount rejects admin/org_admin accounts with
	// ErrCodeAccountTypeNotAllowedInGroup — see port.GroupService.AddAccount.
	g.POST("/:id/accounts/:accountId", func(c *gin.Context) {
		err := groups.AddAccount(c.Request.Context(), model.ID(c.Param("id")), model.ID(c.Param("accountId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.DELETE("/:id/accounts/:accountId", func(c *gin.Context) {
		err := groups.RemoveAccount(c.Request.Context(), model.ID(c.Param("id")), model.ID(c.Param("accountId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	// AddVehicle moves vehicleId into this group, out of whichever group
	// (if any) it previously belonged to — a vehicle belongs to at most one
	// group at a time. A group's vehicles are listed via the ordinary
	// GET /vehicles?groupId=, not a route here — see
	// port.GroupService.AddVehicle.
	g.POST("/:id/vehicles/:vehicleId", func(c *gin.Context) {
		err := groups.AddVehicle(c.Request.Context(), model.ID(c.Param("id")), model.ID(c.Param("vehicleId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.DELETE("/:id/vehicles/:vehicleId", func(c *gin.Context) {
		err := groups.RemoveVehicle(c.Request.Context(), model.ID(c.Param("id")), model.ID(c.Param("vehicleId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})
}
