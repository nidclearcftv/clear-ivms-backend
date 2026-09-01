package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func corsConfig(allowedOrigins []string) cors.Config {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = allowedOrigins
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	cfg.AllowCredentials = false
	cfg.MaxAge = 12 * time.Hour
	return cfg
}

// maxRequestBodyMiddleware rejects request bodies larger than maxBytes,
// guarding the server against unbounded-body requests.
func maxRequestBodyMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// recoveryMiddleware recovers panics from downstream handlers, logs them via
// the adapter's structured logger (instead of gin's default stdout writer),
// and responds 500 rather than crashing the process.
func recoveryMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorw("panic recovered", "error", r, "path", c.Request.URL.Path)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func requestLoggerMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.Infow("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
			"clientIP", c.ClientIP(),
		)
	}
}
