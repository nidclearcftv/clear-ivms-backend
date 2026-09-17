package httpapi

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

// registerObjectRoutes registers generic object read/write routes under
// rg — this is what adapter/storage/local, which has no real external URL
// of its own, points its PutURL/GetURL at (adapter/storage/s3 points
// those at real S3 presigned URLs instead and needs none of this) so
// every adapter honors the same "operations return a URL" contract from
// port.ObjectStorage's callers' point of view (see
// EquipmentModelService's SetPicture/GetPictureURL) — the system's
// upload/download pattern is always a redirect, this backend's bytes
// never involved, regardless of which adapter is active.
//
// Only registered when Options.ObjectStorage is set (see server.go) —
// main.go does that only for the "local" driver.
//
// Gated by authMiddleware alone, not by organization/role membership: the
// key is the only credential, the same way a bearer of a presigned URL
// needs no further authorization from whatever issued it. Keys are
// unguessable (crypto/rand-generated — see EquipmentModelService), so this
// is the same trust model a real presigned URL has, just without an
// expiring signature; anyone able to reach this backend still needs a
// valid session to reach these routes at all.
func registerObjectRoutes(rg *gin.RouterGroup, storage port.ObjectStorage, accounts port.AccountService) {
	g := rg.Group("/objects", authMiddleware(accounts))

	g.PUT("/:key", func(c *gin.Context) {
		contentType := c.ContentType()
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		if err := storage.Put(c.Request.Context(), c.Param("key"), c.Request.Body, contentType); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.GET("/:key", func(c *gin.Context) {
		reader, contentType, err := storage.Get(c.Request.Context(), c.Param("key"))
		if err != nil {
			RespondError(c, err)
			return
		}
		defer reader.Close()

		c.Header("Content-Type", contentType)
		if _, err := io.Copy(c.Writer, reader); err != nil {
			c.Errors = append(c.Errors, &gin.Error{Err: err, Type: gin.ErrorTypePrivate})
		}
	})
}
