package model

// RecaptchaResult is the outcome of verifying a reCAPTCHA response token
// against Google's siteverify endpoint. Score is nil for a v2 (checkbox)
// token, which carries no risk score — only a v3 (invisible) token does.
type RecaptchaResult struct {
	Success bool
	Score   *float64
}
