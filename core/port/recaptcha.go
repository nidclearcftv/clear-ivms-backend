package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// RecaptchaVerifier verifies a reCAPTCHA response token against Google's
// siteverify endpoint. Implemented by adapter/recaptcha.
type RecaptchaVerifier interface {
	// Verify checks token against secretKey. secretKey is passed in per
	// call, rather than fixed at construction, because a v3 (invisible)
	// token and a v2 (checkbox) token are verified against two different
	// secrets — see adapter/http's RecaptchaOptions.
	Verify(ctx context.Context, secretKey, token string) (model.RecaptchaResult, error)
}
