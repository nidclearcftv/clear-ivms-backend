// Package local is a filesystem-backed object storage adapter,
// implementing port.ObjectStorage against a directory on disk. Meant for
// local development/single-instance deployments — see adapter/storage/s3
// (planned) for a deployment spanning multiple instances/hosts.
package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type Options struct {
	// BasePath is the directory objects are stored under. Created
	// (including any missing parents) if it doesn't already exist.
	BasePath string `validate:"required"`

	// PublicBaseURL is this backend's own externally-reachable API base
	// URL (e.g. "http://localhost:8080/api/v1", or the deployed
	// equivalent) — PutURL/GetURL build their URLs by appending
	// "/objects/<key>" to it. It has to be the real external address (not
	// just "localhost" in a deployed environment) because callers (e.g. a
	// browser) use it directly, not through whatever this process
	// otherwise listens on. Requires registerObjectRoutes to actually be
	// registered at that path — see httpapi.Options.ObjectStorage.
	PublicBaseURL string `validate:"required,url"`
}

// Storage implements port.ObjectStorage against BasePath. Each object is
// stored as two files: the raw bytes at BasePath/<key>, and its content
// type (plain text) at BasePath/<key>.contenttype — a plain filesystem has
// no native place to attach metadata to a file, so this sidecar file is
// it.
//
// Unlike a real object store, a file on local disk has no URL a client
// could fetch/PUT it through directly — so PutURL/GetURL point back at
// this same backend's own /objects/:key routes (see registerObjectRoutes)
// instead of somewhere external. That keeps this adapter honoring the
// same "operations return a URL" contract as adapter/storage/s3 will,
// rather than being a permanent exception to it.
type Storage struct {
	basePath      string
	publicBaseURL string
}

func NewStorage(opts Options) (*Storage, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(opts.BasePath, 0o755); err != nil {
		return nil, fmt.Errorf("storage/local: failed to create base directory: %w", err)
	}

	return &Storage{
		basePath:      opts.BasePath,
		publicBaseURL: strings.TrimSuffix(opts.PublicBaseURL, "/"),
	}, nil
}

// objectURL builds the URL registerObjectRoutes serves key at.
func (s *Storage) objectURL(key string) string {
	return s.publicBaseURL + "/objects/" + key
}

func (s *Storage) objectPath(key string) string {
	return filepath.Join(s.basePath, key)
}

func (s *Storage) contentTypePath(key string) string {
	return filepath.Join(s.basePath, key+".contenttype")
}

func (s *Storage) Put(_ context.Context, key string, r io.Reader, contentType string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	f, err := os.Create(s.objectPath(key))
	if err != nil {
		return fmt.Errorf("storage/local: failed to create object: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("storage/local: failed to write object: %w", err)
	}

	if err := os.WriteFile(s.contentTypePath(key), []byte(contentType), 0o644); err != nil {
		return fmt.Errorf("storage/local: failed to write object content type: %w", err)
	}

	return nil
}

// PutURL returns a URL under this backend's own /objects/:key routes
// (see registerObjectRoutes) that a client can PUT key's bytes to
// directly — contentType is unused here (the client sets its own
// Content-Type header on that request, which is what the route actually
// records), but kept in the signature to match port.ObjectStorage, where
// an adapter with a real signing mechanism (e.g. S3) would need it.
func (s *Storage) PutURL(_ context.Context, key string, _ string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	return s.objectURL(key), nil
}

func (s *Storage) Get(_ context.Context, key string) (io.ReadCloser, string, error) {
	if err := validateKey(key); err != nil {
		return nil, "", err
	}

	f, err := os.Open(s.objectPath(key))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, "", model.NewError(model.ErrCodeObjectNotFound, err)
		}
		return nil, "", fmt.Errorf("storage/local: failed to open object: %w", err)
	}

	contentType := "application/octet-stream"
	if data, err := os.ReadFile(s.contentTypePath(key)); err == nil {
		contentType = string(data)
	}

	return f, contentType, nil
}

// GetURL returns a URL under this backend's own /objects/:key routes
// (see registerObjectRoutes) that a client can fetch key's bytes from
// directly.
func (s *Storage) GetURL(_ context.Context, key string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	return s.objectURL(key), nil
}

func (s *Storage) Delete(_ context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	if err := os.Remove(s.objectPath(key)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("storage/local: failed to delete object: %w", err)
	}
	if err := os.Remove(s.contentTypePath(key)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("storage/local: failed to delete object content type: %w", err)
	}

	return nil
}

// validateKey is a defensive check against path traversal — keys are
// always server-generated (see EquipmentModelService), never taken
// directly from untrusted input, so this should never actually trip.
func validateKey(key string) error {
	if key == "" || strings.ContainsAny(key, "/\\") || strings.Contains(key, "..") {
		return fmt.Errorf("storage/local: invalid object key %q", key)
	}
	return nil
}

var _ port.ObjectStorage = (*Storage)(nil)
