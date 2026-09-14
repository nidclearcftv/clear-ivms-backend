package httpapi

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

type createAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password" binding:"required,min=8"`
	Type        string `json:"type" binding:"required,oneof=admin org_admin user"`
	Blocked     bool   `json:"blocked"`
}

type updateAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phoneNumber"`
	Type        string `json:"type" binding:"required,oneof=admin org_admin user"`
	Blocked     bool   `json:"blocked"`
}

type changeAccountPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

// hashPassword bcrypt-hashes a plaintext password for storage — the same
// primitive core/service/seed.go uses, just needed here too since
// AccountService.Create/SetPassword deliberately only accept an
// already-hashed password (see their doc comments): hashing a plaintext
// admin-supplied password is an adapter-layer concern, not the service's.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// registerAccountRoutes registers the admin-only /accounts CRUD, plus
// password reset, session management, and organization-membership listing
// for a given account. requireRolesMiddleware is called with no allowed
// roles, so only model.AccountTypeAdmin — which always bypasses the check —
// can reach any of them. organizations is optional: GET /:id/organizations
// is only registered when it's set (mirrors how server.go treats
// OrganizationService as optional everywhere else).
func registerAccountRoutes(rg *gin.RouterGroup, accounts port.AccountService, organizations port.OrganizationService) {
	g := rg.Group("/accounts", authMiddleware(accounts), requireRolesMiddleware())

	g.POST("", func(c *gin.Context) {
		var req createAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			Fail(c, model.ErrCodeUnknown, err.Error())
			return
		}

		account, err := accounts.Create(c.Request.Context(), model.Account{
			Name:        req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
			Type:        model.AccountType(req.Type),
			Blocked:     req.Blocked,
		}, passwordHash)
		if err != nil {
			RespondError(c, err)
			return
		}
		Created(c, newAccountDTO(account))
	})

	g.GET("", func(c *gin.Context) {
		var filters model.AccountFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := accounts.List(c.Request.Context(), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountListDTO(list))
	})

	g.GET("/:id", func(c *gin.Context) {
		account, err := accounts.Get(c.Request.Context(), model.ID(c.Param("id")))
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountDTO(account))
	})

	g.PUT("/:id", func(c *gin.Context) {
		var req updateAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		account, err := accounts.Update(c.Request.Context(), model.Account{
			ID:          model.ID(c.Param("id")),
			Name:        req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
			Type:        model.AccountType(req.Type),
			Blocked:     req.Blocked,
		})
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountDTO(account))
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := accounts.Delete(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.PUT("/:id/password", func(c *gin.Context) {
		var req changeAccountPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			Fail(c, model.ErrCodeUnknown, err.Error())
			return
		}

		if err := accounts.SetPassword(c.Request.Context(), model.ID(c.Param("id")), passwordHash); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.GET("/:id/sessions", func(c *gin.Context) {
		var filters model.AccountSessionFilters
		if err := c.ShouldBindQuery(&filters); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		list, err := accounts.ListSessions(c.Request.Context(), model.ID(c.Param("id")), filters)
		if err != nil {
			RespondError(c, err)
			return
		}
		OK(c, newAccountSessionListDTO(list))
	})

	g.DELETE("/:id/sessions", func(c *gin.Context) {
		if err := accounts.RevokeAllSessions(c.Request.Context(), model.ID(c.Param("id"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	g.DELETE("/:id/sessions/:sessionId", func(c *gin.Context) {
		if err := accounts.RevokeSession(c.Request.Context(), model.ID(c.Param("sessionId"))); err != nil {
			RespondError(c, err)
			return
		}
		OK(c, nil)
	})

	if organizations != nil {
		g.GET("/:id/organizations", func(c *gin.Context) {
			list, err := organizations.ListFromAccount(c.Request.Context(), model.ID(c.Param("id")))
			if err != nil {
				RespondError(c, err)
				return
			}
			OK(c, newOrganizationSummaryListDTO(list))
		})
	}
}
