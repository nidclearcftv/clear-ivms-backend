package httpapi

import "github.com/gin-gonic/gin"

type statusData struct {
	Status string `json:"status"`
}

func registerStatusRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		OK(c, statusData{Status: "ok"})
	})
}
