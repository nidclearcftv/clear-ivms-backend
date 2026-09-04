package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
)

type createVehicleRequest struct {
	IVMSType    string `json:"ivmsType" binding:"required,oneof=cmsv6"`
	ExternalID  string `json:"externalId" binding:"required"`
	PlateNumber string `json:"plateNumber" binding:"required"`
	GroupID     string `json:"groupId"`
}

type updateVehicleRequest struct {
	IVMSType    string `json:"ivmsType" binding:"required,oneof=cmsv6"`
	ExternalID  string `json:"externalId" binding:"required"`
	PlateNumber string `json:"plateNumber" binding:"required"`
	GroupID     string `json:"groupId"`
}

// registerVehicleRoutes registers /vehicles under rg, gated by
// requireOrganizationMiddleware (mandatory "?orgId="; see that middleware
// for the membership check it does) and
// requireRolesMiddleware(model.AccountTypeOrgAdmin) — so besides
// model.AccountTypeAdmin, which always bypasses role checks, an org_admin
// of the request's organization can reach these routes too; a plain user
// cannot.
func registerVehicleRoutes(rg *gin.RouterGroup, vehicles port.VehicleService, accounts port.AccountService) {
	g := rg.Group("/vehicles",
		authMiddleware(accounts),
		requireRolesMiddleware(model.AccountTypeOrgAdmin),
		requireOrganizationMiddleware(accounts),
	)

	g.POST("", func(c *gin.Context) {
		var req createVehicleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		vehicle, err := vehicles.Create(c.Request.Context(), model.Vehicle{
			OrganizationID: utils.OrganizationID(c.Request.Context()),
			GroupID:        nullableIDFromRequest(req.GroupID),
			IVMSType:       model.IVMSTypeFromString(req.IVMSType),
			ExternalID:     req.ExternalID,
			PlateNumber:    req.PlateNumber,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newVehicleDTO(vehicle))
	})

	g.GET("", func(c *gin.Context) {
		var filters model.VehicleFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := vehicles.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleListDTO(list))
	})

	g.GET("/:id", func(c *gin.Context) {
		vehicle, err := vehicles.Get(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleDTO(vehicle))
	})

	g.PUT("/:id", func(c *gin.Context) {
		var req updateVehicleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		vehicle, err := vehicles.Update(c.Request.Context(), model.Vehicle{
			ID:             model.ID(c.Param("id")),
			OrganizationID: utils.OrganizationID(c.Request.Context()),
			GroupID:        nullableIDFromRequest(req.GroupID),
			IVMSType:       model.IVMSTypeFromString(req.IVMSType),
			ExternalID:     req.ExternalID,
			PlateNumber:    req.PlateNumber,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleDTO(vehicle))
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := vehicles.Delete(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})
}
