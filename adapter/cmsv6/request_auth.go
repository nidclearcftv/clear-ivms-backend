package cmsv6

import (
	"context"
	"fmt"
)

// Login authenticates against /login/loginMobile.do using the configured
// Options.Username/Options.Password and stores the resulting session so
// subsequent requests are authenticated automatically. Callers don't need to
// call this explicitly — it happens lazily on the first authenticated
// request — but calling it eagerly (e.g. at startup) is useful to fail fast
// on bad credentials.
func (s *Server) Login(ctx context.Context) (*LoginResponse, error) {
	return s.login(ctx)
}

func (s *Server) login(ctx context.Context) (*LoginResponse, error) {
	s.log.Info("logging in")

	var result LoginResponse

	resp, err := s.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"account":  s.opts.Username,
			"password": s.opts.Password,
		}).
		SetResult(&result).
		Get("/login/loginMobile.do")
	if err != nil {
		s.log.Errorw("login request failed", "error", err)
		return nil, fmt.Errorf("cmsv6: login request failed: %w", err)
	}

	if resp.IsError() {
		s.log.Errorw("login failed", "status", resp.StatusCode())
		return nil, fmt.Errorf("cmsv6: login failed with status %d", resp.StatusCode())
	}

	if result.Result != 0 {
		s.log.Errorw("login failed", "resultCode", result.Result)
		return nil, fmt.Errorf("cmsv6: login failed with result code %d", result.Result)
	}

	s.setSession(result.Session)
	s.log.Infow("login succeeded", "session", maskSession(result.Session))

	return &result, nil
}
