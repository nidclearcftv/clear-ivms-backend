package port

import "context"

// Cache is the driven (secondary) port for a simple key/value cache. It is
// implemented by outbound adapters, e.g. adapter/cache/memory.
type Cache interface {
	// Get returns the cached value for key, and whether it was found (and
	// not expired).
	Get(ctx context.Context, key string) (value any, found bool)

	// Set stores value under key. Implementations apply their own
	// configured expiration; there is no per-call override for now.
	Set(ctx context.Context, key string, value any)
}
