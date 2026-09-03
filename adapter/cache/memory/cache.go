// Package memory is an in-process cache adapter backed by
// github.com/patrickmn/go-cache, implementing port.Cache.
package memory

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type Options struct {
	// DefaultExpiration is applied to every entry set via Cache.Set — the
	// only expiration supported for now, with no per-key override.
	DefaultExpiration time.Duration `validate:"required,gt=0"`

	// CleanupInterval controls how often expired entries are purged from
	// memory in the background. Defaults to DefaultExpiration when unset.
	CleanupInterval time.Duration `validate:"omitempty,gt=0"`
}

// Cache is an in-process, single-instance cache. It implements port.Cache.
type Cache struct {
	cache *gocache.Cache
}

func NewCache(opts Options) (*Cache, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	cleanupInterval := opts.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = opts.DefaultExpiration
	}

	return &Cache{
		cache: gocache.New(opts.DefaultExpiration, cleanupInterval),
	}, nil
}

// Get returns the cached value for key, and whether it was found (and not
// expired).
func (c *Cache) Get(_ context.Context, key string) (any, bool) {
	return c.cache.Get(key)
}

// Set stores value under key using the cache's configured
// Options.DefaultExpiration.
func (c *Cache) Set(_ context.Context, key string, value any) {
	c.cache.SetDefault(key, value)
}

// Del removes key from the cache. A no-op if key isn't present.
func (c *Cache) Del(_ context.Context, key string) {
	c.cache.Delete(key)
}

var _ port.Cache = (*Cache)(nil)
