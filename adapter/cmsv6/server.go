package cmsv6

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type Options struct {
	Context                context.Context
	Logger                 *zap.SugaredLogger
	URL                    string        `validate:"required,url"`
	Username               string        `validate:"required"`
	Password               string        `validate:"required"`
	RefreshSessionInterval time.Duration `validate:"omitempty,gt=0"`
}

type Server struct {
	client *resty.Client
	opts   Options
	log    *zap.SugaredLogger

	sessionMu sync.RWMutex
	session   string

	loginMu sync.Mutex

	cancel context.CancelFunc
	done   chan struct{}
}

func NewServer(opts Options) (*Server, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	parent := opts.Context
	if parent == nil {
		parent = context.Background()
	}

	log := opts.Logger
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	log = log.With("adapter", "cmsv6", "account", opts.Username)

	client := resty.New().
		SetBaseURL(opts.URL)

	s := &Server{
		client: client,
		opts:   opts,
		log:    log,
	}

	client.OnBeforeRequest(func(_ *resty.Client, r *resty.Request) error {
		if session := s.currentSession(); session != "" {
			r.SetCookie(&http.Cookie{Name: "JSESSIONID", Value: session})
		}
		return nil
	})

	if opts.RefreshSessionInterval > 0 {
		ctx, cancel := context.WithCancel(parent)
		s.cancel = cancel
		s.done = make(chan struct{})

		log.Infow("starting background session refresh", "interval", opts.RefreshSessionInterval)
		go s.refreshSessionLoop(ctx)
	}

	return s, nil
}

func (s *Server) Start() error {
	if err := s.ensureSession(s.opts.Context); err != nil {
		return err
	}
	return nil
}

// Close stops the background session refresh started by NewServer. It is a
// no-op if Options.RefreshSessionInterval was not set.
func (s *Server) Close() {
	if s.cancel == nil {
		return
	}
	s.log.Info("stopping background session refresh")
	s.cancel()
	<-s.done
}

func (s *Server) refreshSessionLoop(ctx context.Context) {
	defer close(s.done)

	ticker := time.NewTicker(s.opts.RefreshSessionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("background session refresh stopped")
			return
		case <-ticker.C:
			s.log.Debug("refreshing session")
			// Best-effort: a failed refresh leaves the current session in
			// place (still possibly valid), and ensureSession will retry
			// on the next authenticated request if it has since expired.
			if _, err := s.login(ctx); err != nil {
				s.log.Warnw("session refresh failed, keeping previous session", "error", err)
			}
		}
	}
}

func (s *Server) currentSession() string {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	return s.session
}

func (s *Server) setSession(session string) {
	s.sessionMu.Lock()
	s.session = session
	s.sessionMu.Unlock()
}

// ensureSession lazily logs in on first use of an authenticated request.
// Concurrent callers block on loginMu so only one login happens at a time.
func (s *Server) ensureSession(ctx context.Context) error {
	if s.currentSession() != "" {
		return nil
	}

	s.loginMu.Lock()
	defer s.loginMu.Unlock()

	if s.currentSession() != "" {
		return nil
	}

	s.log.Debug("no active session, logging in")
	_, err := s.login(ctx)
	return err
}

// maskSession trims a session token down to a form safe to put in logs.
func maskSession(session string) string {
	if len(session) <= 8 {
		return "***"
	}
	return session[:4] + "…" + session[len(session)-4:]
}
