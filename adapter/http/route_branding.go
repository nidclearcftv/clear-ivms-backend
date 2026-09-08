package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

type createBrandingRequest struct {
	Name   string     `json:"name" binding:"required"`
	Domain string     `json:"domain" binding:"required"`
	Config model.JSON `json:"config"`
}

type updateBrandingRequest struct {
	Name   string     `json:"name" binding:"required"`
	Domain string     `json:"domain" binding:"required"`
	Config model.JSON `json:"config"`
}

// registerPublicBrandingRoutes registers GET /branding?domain=, an
// unauthenticated lookup clients use to fetch a domain's branding config
// before any account is involved. Unlike every other resource's routes,
// this one does not sit behind authMiddleware, so it's registered
// independently of whether AccountService is configured — see server.go.
func registerPublicBrandingRoutes(rg *gin.RouterGroup, brandings port.BrandingService) {
	rg.GET("/branding", func(c *gin.Context) {
		domain := c.Query("domain")
		if domain == "" {
			Fail(c, model.ErrCodeInvalidRequest, "missing required query parameter: domain")
			return
		}

		branding, err := brandings.GetByDomain(c.Request.Context(), domain)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newBrandingDTO(branding))
	})
}

// registerBrandingRoutes registers the admin-only /brandings CRUD:
// requireRolesMiddleware is called with no allowed roles, so only
// model.AccountTypeAdmin — which always bypasses the check — can reach any
// of them.
func registerBrandingRoutes(rg *gin.RouterGroup, brandings port.BrandingService, accounts port.AccountService) {
	g := rg.Group("/brandings", authMiddleware(accounts), requireRolesMiddleware())

	g.POST("", func(c *gin.Context) {
		var req createBrandingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		branding, err := brandings.Create(c.Request.Context(), model.Branding{
			Name:   req.Name,
			Domain: req.Domain,
			Config: req.Config,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newBrandingDTO(branding))
	})

	g.GET("", func(c *gin.Context) {
		var filters model.BrandingFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := brandings.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newBrandingListDTO(list))
	})

	g.GET("/:id", func(c *gin.Context) {
		branding, err := brandings.Get(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newBrandingDTO(branding))
	})

	g.PUT("/:id", func(c *gin.Context) {
		var req updateBrandingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		branding, err := brandings.Update(c.Request.Context(), model.Branding{
			ID:     model.ID(c.Param("id")),
			Name:   req.Name,
			Domain: req.Domain,
			Config: req.Config,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newBrandingDTO(branding))
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := brandings.Delete(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})
}
