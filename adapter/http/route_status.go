package httpapi

import "github.com/gin-gonic/gin"

func registerStatusRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		OK(c, nil)
	})
}
