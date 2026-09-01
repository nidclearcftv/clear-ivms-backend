package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusResponse struct {
	Status string `json:"status"`
}

func registerStatusRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, statusResponse{Status: "ok"})
	})
}
