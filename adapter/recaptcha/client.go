// Package recaptcha implements port.RecaptchaVerifier against Google's
// reCAPTCHA siteverify endpoint.
package recaptcha

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

const siteverifyURL = "https://www.google.com/recaptcha/api/siteverify"

// Client implements port.RecaptchaVerifier.
type Client struct {
	http *resty.Client
}

func NewClient() *Client {
	return &Client{http: resty.New()}
}

type siteverifyResponse struct {
	Success    bool     `json:"success"`
	Score      *float64 `json:"score"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify calls Google's siteverify endpoint. A network/transport failure or
// a non-2xx response from Google is returned as a plain error (an
// infrastructure problem, not a caller-input one); a token that Google
// itself rejects (missing, malformed, expired, already used, ...) is
// reported through RecaptchaResult.Success, not an error.
func (c *Client) Verify(ctx context.Context, secretKey, token string) (model.RecaptchaResult, error) {
	var result siteverifyResponse
	resp, err := c.http.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"secret":   secretKey,
			"response": token,
		}).
		SetResult(&result).
		Post(siteverifyURL)
	if err != nil {
		return model.RecaptchaResult{}, fmt.Errorf("recaptcha: failed to call siteverify: %w", err)
	}
	if resp.IsError() {
		return model.RecaptchaResult{}, fmt.Errorf("recaptcha: siteverify returned status %d", resp.StatusCode())
	}

	return model.RecaptchaResult{Success: result.Success, Score: result.Score}, nil
}

var _ port.RecaptchaVerifier = (*Client)(nil)
