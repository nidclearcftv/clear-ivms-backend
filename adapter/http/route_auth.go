package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

// sessionCookieName is the cookie the session token is stored in. It's
// HttpOnly (never readable from JS) and Secure (browser-sent over HTTPS
// only) — disable Secure only for local plain-HTTP testing.
const sessionCookieName = "session_token"

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func registerAuthRoutes(rg *gin.RouterGroup, accounts port.AccountService) {
	rg.POST("/login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		account, token, err := accounts.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			RespondError(c, err)
			return
		}

		setSessionCookie(c, token)
		OK(c, newAccountDTO(account))
	})

	// Logout always clears the cookie client-side, regardless of whether a
	// session existed or the revoke below fully succeeds server-side — the
	// client asked to be logged out, so it shouldn't keep holding a token.
	rg.POST("/logout", func(c *gin.Context) {
		token, cookieErr := c.Cookie(sessionCookieName)
		clearSessionCookie(c)

		if cookieErr != nil {
			OK(c, nil)
			return
		}

		if err := accounts.Logout(c.Request.Context(), token); err != nil {
			var merr *model.Error
			if errors.As(err, &merr) && merr.Code == model.ErrCodeAccountSessionNotFound {
				OK(c, nil)
				return
			}
			RespondError(c, err)
			return
		}

		OK(c, nil)
	})

	rg.GET("/me", authMiddleware(accounts), func(c *gin.Context) {
		account, ok := contextAccount(c)
		if !ok {
			// authMiddleware always sets this before calling c.Next() on
			// success, so reaching here means it didn't run at all.
			Fail(c, model.ErrCodeUnknown)
			return
		}
		OK(c, newAccountDTO(account))
	})
}

func setSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
