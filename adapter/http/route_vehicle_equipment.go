package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

type createVehicleEquipmentRequest struct {
	EquipmentModelID string `json:"equipmentModelId" binding:"required"`
	SerialNumber     string `json:"serialNumber"`
	Description      string `json:"description"`
}

// updateVehicleEquipmentRequest deliberately excludes EquipmentModelID —
// see VehicleEquipmentService.Update, which always preserves the
// registration's existing equipment model (and the vehicle it belongs
// to) regardless of what's passed in the model.VehicleEquipment it's
// given.
type updateVehicleEquipmentRequest struct {
	SerialNumber string `json:"serialNumber"`
	Description  string `json:"description"`
}

// registerVehicleEquipmentRoutes registers /vehicles/:id/equipment under
// rg, gated by requireOrganizationMiddleware (mandatory "?orgId="; see
// that middleware for the membership check it does) and
// requireRolesMiddleware(model.AccountTypeOrgAdmin) — so besides
// model.AccountTypeAdmin, which always bypasses role checks, an org_admin
// of the request's organization can reach these routes too; a plain user
// cannot.
//
// The vehicle segment is named :id, not :vehicleId, because
// registerVehicleRoutes already registered /vehicles/:id — gin's router
// panics if two registrations under the same prefix name that wildcard
// segment differently. The registration's own id is :equipmentId instead,
// to avoid colliding with that same name at a different depth.
func registerVehicleEquipmentRoutes(rg *gin.RouterGroup, vehicleEquipment port.VehicleEquipmentService, accounts port.AccountService) {
	g := rg.Group("/vehicles/:id/equipment",
		authMiddleware(accounts),
		requireRolesMiddleware(model.AccountTypeOrgAdmin),
		requireOrganizationMiddleware(accounts),
	)

	g.POST("", func(c *gin.Context) {
		var req createVehicleEquipmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		ve, err := vehicleEquipment.Create(c.Request.Context(), model.VehicleEquipment{
			VehicleID:        model.ID(c.Param("id")),
			EquipmentModelID: model.ID(req.EquipmentModelID),
			SerialNumber:     req.SerialNumber,
			Description:      req.Description,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newVehicleEquipmentDTO(ve))
	})

	g.GET("", func(c *gin.Context) {
		items, err := vehicleEquipment.List(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleEquipmentListDTO(items))
	})

	g.GET("/:equipmentId", func(c *gin.Context) {
		ve, err := vehicleEquipment.Get(c.Request.Context(), model.ID(c.Param("equipmentId")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleEquipmentDTO(ve))
	})

	g.PUT("/:equipmentId", func(c *gin.Context) {
		var req updateVehicleEquipmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		ve, err := vehicleEquipment.Update(c.Request.Context(), model.VehicleEquipment{
			ID:           model.ID(c.Param("equipmentId")),
			SerialNumber: req.SerialNumber,
			Description:  req.Description,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleEquipmentDTO(ve))
	})

	g.DELETE("/:equipmentId", func(c *gin.Context) {
		if err := vehicleEquipment.Delete(c.Request.Context(), model.ID(c.Param("equipmentId"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})
}
