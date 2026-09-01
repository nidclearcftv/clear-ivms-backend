package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

func registerVehicleRoutes(rg *gin.RouterGroup, vehicles port.VehicleService) {
	rg.GET("/vehicles", func(c *gin.Context) {
		var filters model.VehicleFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := vehicles.ListVehicles(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newVehicleListDTO(list))
	})
}
