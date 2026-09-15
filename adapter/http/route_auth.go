package httpapi

import (
	"context"
	"errors"
	"fmt"
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

// RecaptchaOptions gates POST /login with reCAPTCHA verification.
// Verification happens entirely at the HTTP layer (not inside
// AccountService/port.AccountService) since it's a request-level
// precondition on the login attempt, not account domain logic — the same
// reasoning that keeps AllowInsecureCookies (Options' other login-adjacent
// setting) out of the account service too.
type RecaptchaOptions struct {
	Enabled bool

	// Verifier calls Google's siteverify endpoint. Required when Enabled.
	Verifier port.RecaptchaVerifier

	// V3SecretKey/V2SecretKey are the secret keys for the invisible (v3)
	// and checkbox (v2) reCAPTCHA site keys respectively — two different
	// reCAPTCHA products, each with its own site/secret key pair. Both
	// required when Enabled.
	V3SecretKey string
	V2SecretKey string

	// ScoreThreshold is the minimum v3 score (0-1) accepted without a
	// step-up v2 challenge. Below it, /login responds with
	// ErrCodeRecaptchaChallengeRequired instead of completing the login,
	// so the client can render the v2 checkbox and retry with
	// RecaptchaChallengeToken. Defaults to 0.5 (see NewServer).
	ScoreThreshold float64
}

type loginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`

	// RecaptchaToken is the v3 (invisible) token, sent on every login
	// attempt while reCAPTCHA is enabled. RecaptchaChallengeToken is the
	// v2 (checkbox) token, sent only on a retry after a previous attempt
	// returned ErrCodeRecaptchaChallengeRequired. See verifyRecaptcha.
	RecaptchaToken          string `json:"recaptchaToken"`
	RecaptchaChallengeToken string `json:"recaptchaChallengeToken"`
}

// verifyRecaptcha enforces recaptcha on a login attempt, or does nothing
// if it's disabled. A RecaptchaChallengeToken (the v2 checkbox, completed
// as a step-up after a previous attempt returned
// ErrCodeRecaptchaChallengeRequired) is checked first and, if present, is
// the only thing verified — success always allows the login through, with
// no further scoring (v2 tokens carry no score). Otherwise the v3
// (invisible) token is verified and its score decides whether the login
// can proceed immediately or must be retried with a completed v2
// challenge: a low score never hard-blocks the attempt by itself, only
// requires the extra step, since v3 scores alone aren't reliable enough to
// permanently lock out a real user.
func verifyRecaptcha(ctx context.Context, recaptcha RecaptchaOptions, req loginRequest) error {
	if !recaptcha.Enabled {
		return nil
	}

	if req.RecaptchaChallengeToken != "" {
		result, err := recaptcha.Verifier.Verify(ctx, recaptcha.V2SecretKey, req.RecaptchaChallengeToken)
		if err != nil {
			return fmt.Errorf("recaptcha: failed to verify challenge token: %w", err)
		}
		if !result.Success {
			return model.NewError(model.ErrCodeRecaptchaFailed, nil)
		}
		return nil
	}

	if req.RecaptchaToken == "" {
		return model.NewError(model.ErrCodeRecaptchaFailed, nil)
	}

	result, err := recaptcha.Verifier.Verify(ctx, recaptcha.V3SecretKey, req.RecaptchaToken)
	if err != nil {
		return fmt.Errorf("recaptcha: failed to verify token: %w", err)
	}
	if !result.Success {
		return model.NewError(model.ErrCodeRecaptchaFailed, nil)
	}
	if result.Score == nil || *result.Score < recaptcha.ScoreThreshold {
		return model.NewError(model.ErrCodeRecaptchaChallengeRequired, nil)
	}

	return nil
}

func registerAuthRoutes(rg *gin.RouterGroup, accounts port.AccountService, organizations port.OrganizationService, cookieSecure bool, recaptcha RecaptchaOptions) {
	rg.POST("/login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, model.ErrCodeInvalidRequest, err.Error())
			return
		}

		if err := verifyRecaptcha(c.Request.Context(), recaptcha, req); err != nil {
			RespondError(c, err)
			return
		}

		account, token, expiresAt, err := accounts.Login(
			c.Request.Context(),
			req.Email,
			req.Password,
			req.RememberMe,
			c.Request.UserAgent(),
			c.ClientIP(),
		)
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
		if organizations != nil && account.Type != model.AccountTypeAdmin {
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
