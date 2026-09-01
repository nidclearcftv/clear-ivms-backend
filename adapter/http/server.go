// Package httpapi is the HTTP adapter: a gin-based server exposing the
// application's HTTP API under the hardcoded /api/v1 prefix.
package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

const apiV1Prefix = "/api/v1"

const (
	defaultMaxRequestBodyBytes = 1 << 20 // 1 MiB
	defaultReadHeaderTimeout   = 5 * time.Second
	defaultReadTimeout         = 10 * time.Second
	defaultWriteTimeout        = 10 * time.Second
	defaultIdleTimeout         = 60 * time.Second
	defaultShutdownTimeout     = 10 * time.Second
)

type Options struct {
	Logger *zap.SugaredLogger

	Addr string `validate:"required"`

	// VehicleService backs the /api/v1/vehicles route.
	VehicleService port.VehicleService `validate:"required"`

	// AllowedOrigins is the CORS allow-list. Leave empty to disable CORS
	// entirely: cross-origin browser requests are blocked (the safe
	// default), same-origin and non-browser clients are unaffected.
	AllowedOrigins []string `validate:"omitempty,dive,required"`

	// MaxRequestBodyBytes caps request body size to guard against
	// unbounded-body requests. Defaults to 1 MiB.
	MaxRequestBodyBytes int64 `validate:"omitempty,gt=0"`

	ReadHeaderTimeout time.Duration `validate:"omitempty,gt=0"`
	ReadTimeout       time.Duration `validate:"omitempty,gt=0"`
	WriteTimeout      time.Duration `validate:"omitempty,gt=0"`
	IdleTimeout       time.Duration `validate:"omitempty,gt=0"`

	// ShutdownTimeout bounds how long Close waits for in-flight requests to
	// finish. Defaults to 10s.
	ShutdownTimeout time.Duration `validate:"omitempty,gt=0"`
}

type Server struct {
	engine *gin.Engine
	http   *http.Server
	log    *zap.SugaredLogger
	opts   Options
}

func NewServer(opts Options) (*Server, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	if opts.MaxRequestBodyBytes == 0 {
		opts.MaxRequestBodyBytes = defaultMaxRequestBodyBytes
	}
	if opts.ReadHeaderTimeout == 0 {
		opts.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = defaultReadTimeout
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = defaultWriteTimeout
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = defaultIdleTimeout
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = defaultShutdownTimeout
	}

	log := opts.Logger
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	log = log.With("adapter", "http")

	// Avoid gin's default debug mode: it prints a noisy startup banner and
	// per-route registration logs straight to stdout, bypassing our
	// structured logger.
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(recoveryMiddleware(log))
	engine.Use(requestLoggerMiddleware(log))
	engine.Use(maxRequestBodyMiddleware(opts.MaxRequestBodyBytes))

	if len(opts.AllowedOrigins) > 0 {
		engine.Use(cors.New(corsConfig(opts.AllowedOrigins)))
	}

	// Don't trust any inbound proxy headers (X-Forwarded-For, etc.) unless
	// explicitly configured to sit behind one; gin trusts all proxies by
	// default, which lets clients spoof their own IP.
	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf("http: failed to configure trusted proxies: %w", err)
	}

	v1 := engine.Group(apiV1Prefix)
	registerStatusRoutes(v1)
	registerVehicleRoutes(v1, opts.VehicleService)

	return &Server{
		engine: engine,
		log:    log,
		opts:   opts,
		http: &http.Server{
			Addr:              opts.Addr,
			Handler:           engine,
			ReadHeaderTimeout: opts.ReadHeaderTimeout,
			ReadTimeout:       opts.ReadTimeout,
			WriteTimeout:      opts.WriteTimeout,
			IdleTimeout:       opts.IdleTimeout,
		},
	}, nil
}

// Start binds the configured address and begins serving in the background.
// It returns once the listener is bound, so bind errors (e.g. port already
// in use) surface immediately instead of only inside the goroutine.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return fmt.Errorf("http: failed to listen on %s: %w", s.opts.Addr, err)
	}

	go func() {
		s.log.Infow("http server listening", "addr", s.opts.Addr)
		if err := s.http.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Errorw("http server stopped unexpectedly", "error", err)
		}
	}()

	return nil
}

// Close gracefully shuts down the HTTP server, waiting up to
// Options.ShutdownTimeout for in-flight requests to finish.
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
	defer cancel()

	s.log.Info("shutting down http server")
	if err := s.http.Shutdown(ctx); err != nil {
		s.log.Errorw("http server graceful shutdown failed", "error", err)
	}
}
