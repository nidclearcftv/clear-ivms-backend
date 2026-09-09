package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

// sessionCookieName is the cookie the session token is stored in. It's
// always HttpOnly (never readable from JS); whether it's also Secure
// (browser-sent over HTTPS only) is controlled by the cookieSecure param
// threaded through from Options.AllowInsecureCookies — disable Secure only
// for local plain-HTTP testing.
const sessionCookieName = "session_token"

type loginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`
}

func registerAuthRoutes(rg *gin.RouterGroup, accounts port.AccountService, organizations port.OrganizationService, cookieSecure bool) {
	rg.POST("/login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		account, token, expiresAt, err := accounts.Login(c.Request.Context(), req.Email, req.Password, req.RememberMe)
		if err != nil {
			RespondError(c, err)
			return
		}

		setSessionCookie(c, token, cookieSecure, req.RememberMe, expiresAt)
		OK(c, newAccountDTO(account))
	})

	// Logout always clears the cookie client-side, regardless of whether a
	// session existed or the revoke below fully succeeds server-side — the
	// client asked to be logged out, so it shouldn't keep holding a token.
	rg.POST("/logout", func(c *gin.Context) {
		token, cookieErr := c.Cookie(sessionCookieName)
		clearSessionCookie(c, cookieSecure)

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

		var orgs []model.Organization
		if organizations != nil {
			list, err := organizations.ListFromAccount(c.Request.Context(), account.ID)
			if err != nil {
				RespondError(c, err)
				return
			}
			orgs = list.Items
		}

		OK(c, newMeDTO(account, orgs))
	})
}

// setSessionCookie writes the session cookie. When persistent is false
// (the default, no "remember me"), it's a browser-session cookie — cleared
// as soon as the browser closes, regardless of how long the server-side
// session itself remains valid — so a user who didn't ask to be remembered
// has to log in again next time they open the browser. When persistent is
// true, Expires is set to the session's actual server-side expiry so the
// cookie survives browser restarts for exactly as long as the session does.
func setSessionCookie(c *gin.Context, token string, secure bool, persistent bool, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	if persistent {
		cookie.Expires = expiresAt
	}
	http.SetCookie(c.Writer, cookie)
}

func clearSessionCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
