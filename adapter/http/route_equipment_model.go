package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
)

type createEquipmentModelRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required,oneof=primary accessory"`
}

// updateEquipmentModelRequest deliberately excludes Public — see
// EquipmentModelService.Update, which always preserves the equipment
// model's existing value for it regardless of what's passed in the
// model.EquipmentModel it's given. See PUT /:id/public instead, an
// admin-only route.
type updateEquipmentModelRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required,oneof=primary accessory"`
}

type setEquipmentModelPublicRequest struct {
	Public bool `json:"public"`
}

// registerEquipmentModelRoutes registers /equipment-models under rg,
// gated by requireOrganizationMiddleware (mandatory "?orgId="; see that
// middleware for the membership check it does) and
// requireRolesMiddleware(model.AccountTypeOrgAdmin) — so besides
// model.AccountTypeAdmin, which always bypasses role checks, an org_admin
// of the request's organization can reach these routes too; a plain user
// cannot.
//
// PUT /:id/public is stricter than the rest of the group: it additionally
// requires requireRolesMiddleware() with no allowed roles, so only
// model.AccountTypeAdmin can reach it — an org_admin who passes the
// group's own role check above is still rejected there.
func registerEquipmentModelRoutes(rg *gin.RouterGroup, equipmentModels port.EquipmentModelService, accounts port.AccountService) {
	g := rg.Group("/equipment-models",
		authMiddleware(accounts),
		requireRolesMiddleware(model.AccountTypeOrgAdmin),
		requireOrganizationMiddleware(accounts),
	)

	g.POST("", func(c *gin.Context) {
		var req createEquipmentModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		equipmentModel, err := equipmentModels.Create(c.Request.Context(), model.EquipmentModel{
			Name:           req.Name,
			Description:    req.Description,
			Type:           model.EquipmentModelType(req.Type),
			OrganizationID: utils.OrganizationID(c.Request.Context()),
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newEquipmentModelDTO(equipmentModel))
	})

	g.GET("", func(c *gin.Context) {
		var filters model.EquipmentModelFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := equipmentModels.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newEquipmentModelListDTO(list))
	})

	g.GET("/:id", func(c *gin.Context) {
		equipmentModel, err := equipmentModels.Get(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newEquipmentModelDTO(equipmentModel))
	})

	g.PUT("/:id", func(c *gin.Context) {
		var req updateEquipmentModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		equipmentModel, err := equipmentModels.Update(c.Request.Context(), model.EquipmentModel{
			ID:             model.ID(c.Param("id")),
			Name:           req.Name,
			Description:    req.Description,
			Type:           model.EquipmentModelType(req.Type),
			OrganizationID: utils.OrganizationID(c.Request.Context()),
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newEquipmentModelDTO(equipmentModel))
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := equipmentModels.Delete(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	// PUT /:id/public is admin-only — see the requireRolesMiddleware()
	// call added here, on top of the group's own.
	g.PUT("/:id/public", requireRolesMiddleware(), func(c *gin.Context) {
		var req setEquipmentModelPublicRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		if err := equipmentModels.SetPublic(c.Request.Context(), model.ID(c.Param("id")), req.Public); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	// PUT /:id/picture is how a client uploads (or replaces) the
	// equipment model's picture — but this backend's own bytes are never
	// involved: it takes the request's Content-Type header as what the
	// client intends to upload, validates it, and always responds with a
	// redirect to a presigned URL the client must then PUT the actual
	// bytes to directly (see EquipmentModelService.SetPicture — this is
	// the system's one and only upload pattern, with no direct-proxy
	// fallback and no separate confirmation step).
	g.PUT("/:id/picture", func(c *gin.Context) {
		contentType := c.ContentType()
		if !model.EquipmentModelAllowedPictureContentTypes[contentType] {
			Fail(c, model.ErrCodeInvalidRequest, "unsupported picture content type: "+contentType)
			return
		}

		url, err := equipmentModels.SetPicture(c.Request.Context(), model.ID(c.Param("id")), contentType)
		if err != nil {
			RespondError(c, err)
			return
		}
		c.Redirect(http.StatusMovedPermanently, url)
	})

	// GET /:id/picture always redirects to a presigned URL for the
	// equipment model's current picture (see
	// EquipmentModelService.GetPictureURL) — this backend never streams
	// the bytes itself. A public equipment model's picture is readable
	// from any organization the same way its other fields are.
	g.GET("/:id/picture", func(c *gin.Context) {
		url, err := equipmentModels.GetPictureURL(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		c.Redirect(http.StatusMovedPermanently, url)
	})

	g.DELETE("/:id/picture", func(c *gin.Context) {
		equipmentModel, err := equipmentModels.DeletePicture(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newEquipmentModelDTO(equipmentModel))
	})
}
