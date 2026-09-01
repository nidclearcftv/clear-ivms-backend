package utils

import (
	"context"

	"golang.org/x/sync/singleflight"

	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

// fetchGroup dedupes concurrent FetchThrough calls across the whole
// application. It's a single global rather than one per caller because
// dedup keys are already namespaced per resource (e.g. model.VechicleListKey
// prefixes with "vehicle_list:"), so one Group safely serves every caller.
var fetchGroup singleflight.Group

// FetchThrough performs a cached, request-coalesced read: concurrent calls
// sharing key are deduplicated, a cache hit is returned as-is, and a miss
// calls fetch, populates cache with its result, then returns it. For safety,
// if the key is empty, fetch is called directly and the cache is not used.
func FetchThrough[T any](ctx context.Context, cache port.Cache, key string, fetch func() (T, error)) (T, error) {
	if key == "" {
		return fetch()
	}

	v, err, _ := fetchGroup.Do(key, func() (any, error) {
		if cached, ok := cache.Get(ctx, key); ok {
			if value, ok := cached.(T); ok {
				return value, nil
			}
		}

		value, err := fetch()
		if err != nil {
			return nil, err
		}

		cache.Set(ctx, key, value)
		return value, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}

	return v.(T), nil
}
