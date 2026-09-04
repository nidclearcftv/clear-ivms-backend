package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

const defaultSessionTTL = 24 * time.Hour

type AccountServiceOptions struct {
	Repository port.AccountRepository `validate:"required"`
	Cache      port.Cache             `validate:"required"`

	// SessionTTL is how long a session created by Login stays valid.
	// Defaults to 24h.
	SessionTTL time.Duration `validate:"omitempty,gt=0"`
}

// AccountService implements port.AccountService. Most methods are thin
// passthroughs to port.AccountRepository (see VehicleService for why that
// separation exists even so); Login and Logout carry real logic —
// credential verification and session issuance/revocation — that has no
// business living in the repository.
type AccountService struct {
	repo       port.AccountRepository
	cache      port.Cache
	sessionTTL time.Duration
}

func NewAccountService(opts AccountServiceOptions) (*AccountService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	sessionTTL := opts.SessionTTL
	if sessionTTL == 0 {
		sessionTTL = defaultSessionTTL
	}

	return &AccountService{repo: opts.Repository, cache: opts.Cache, sessionTTL: sessionTTL}, nil
}

// Create expects passwordHash to already be hashed by the caller — same
// contract as AccountRepository.Create. There's currently no port method
// that takes a plaintext password and hashes it for account creation; see
// Login for the hashing primitives this service does own.
func (s *AccountService) Create(ctx context.Context, account model.Account, passwordHash string) (model.Account, error) {
	return s.repo.Create(ctx, account, passwordHash)
}

func (s *AccountService) Get(ctx context.Context, id model.ID) (model.Account, error) {
	return s.repo.Get(ctx, id)
}

func (s *AccountService) List(ctx context.Context, filters model.AccountFilters) (model.List[model.Account], error) {
	return s.repo.List(ctx, filters)
}

// Update invalidates the cache entry Authenticate populates for this
// account, so a change (e.g. Blocked) takes effect on the next
// authenticated request instead of waiting out the cache's expiration.
func (s *AccountService) Update(ctx context.Context, account model.Account) (model.Account, error) {
	updated, err := s.repo.Update(ctx, account)
	if err != nil {
		return model.Account{}, err
	}

	s.cache.Del(ctx, model.AccountKey(account.ID))

	return updated, nil
}

// Delete invalidates the cache entry Authenticate populates for this
// account; see Update.
func (s *AccountService) Delete(ctx context.Context, id model.ID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.cache.Del(ctx, model.AccountKey(id))

	return nil
}

func (s *AccountService) GetPassword(ctx context.Context, id model.ID) (string, error) {
	return s.repo.GetPassword(ctx, id)
}

func (s *AccountService) SetPassword(ctx context.Context, id model.ID, passwordHash string) error {
	return s.repo.SetPassword(ctx, id, passwordHash)
}

// Login verifies email/password and, on success, starts a new session. See
// port.AccountService.Login for the exact contract (in particular: unknown
// email and wrong password both fail as ErrCodeInvalidCredentials, never
// distinguished).
func (s *AccountService) Login(ctx context.Context, email, password string) (model.Account, string, error) {
	account, passwordHash, err := s.repo.GetByEmailWithPassword(ctx, email)
	if err != nil {
		var merr *model.Error
		if errors.As(err, &merr) && merr.Code == model.ErrCodeAccountNotFound {
			return model.Account{}, "", model.NewError(model.ErrCodeInvalidCredentials, err)
		}
		return model.Account{}, "", err
	}

	if account.Blocked {
		return model.Account{}, "", model.NewError(model.ErrCodeAccountBlocked, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return model.Account{}, "", model.NewError(model.ErrCodeInvalidCredentials, err)
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return model.Account{}, "", fmt.Errorf("service: failed to generate session token: %w", err)
	}

	if _, err := s.repo.CreateSession(ctx, model.AccountSession{
		AccountID: account.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}); err != nil {
		return model.Account{}, "", err
	}

	return account, token, nil
}

// Logout invalidates the cache entry Authenticate populates for this
// session, so a revoked session stops authenticating on the very next
// request instead of continuing to work until the cache entry expires.
func (s *AccountService) Logout(ctx context.Context, token string) error {
	tokenHash := hashSessionToken(token)

	if err := s.repo.RevokeSessionByToken(ctx, tokenHash); err != nil {
		return err
	}

	s.cache.Del(ctx, model.AccountSessionKey(tokenHash))

	return nil
}

// Authenticate validates a raw session token and returns the account it
// belongs to. See port.AccountService.Authenticate for the exact contract.
//
// The session lookup is cached the same way the account lookup below is
// (Authenticate runs on every authenticated request), keyed by token hash
// since that's what's on hand here. Logout, RevokeSession and
// RevokeAllSessions all invalidate this cache using the token hash(es)
// AccountRepository hands back from the revoke itself, so a revoked
// session stops authenticating immediately rather than surviving until
// the cache entry expires.
func (s *AccountService) Authenticate(ctx context.Context, token string) (model.Account, error) {
	tokenHash := hashSessionToken(token)
	session, err := utils.FetchThrough(ctx, s.cache, model.AccountSessionKey(tokenHash), func() (model.AccountSession, error) {
		return s.repo.GetSession(ctx, tokenHash)
	})
	if err != nil {
		var merr *model.Error
		if errors.As(err, &merr) && merr.Code == model.ErrCodeAccountSessionNotFound {
			return model.Account{}, model.NewError(model.ErrCodeInvalidCredentials, err)
		}
		return model.Account{}, err
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return model.Account{}, model.NewError(model.ErrCodeInvalidCredentials, nil)
	}

	account, err := utils.FetchThrough(ctx, s.cache, model.AccountKey(session.AccountID), func() (model.Account, error) {
		return s.repo.Get(ctx, session.AccountID)
	})
	if err != nil {
		return model.Account{}, err
	}

	if account.Blocked {
		return model.Account{}, model.NewError(model.ErrCodeAccountBlocked, nil)
	}

	return account, nil
}

func (s *AccountService) CreateSession(ctx context.Context, session model.AccountSession) (model.AccountSession, error) {
	return s.repo.CreateSession(ctx, session)
}

func (s *AccountService) GetSession(ctx context.Context, tokenHash string) (model.AccountSession, error) {
	return s.repo.GetSession(ctx, tokenHash)
}

func (s *AccountService) ListSessions(ctx context.Context, accountID model.ID) (model.List[model.AccountSession], error) {
	return s.repo.ListSessions(ctx, accountID)
}

// RevokeSession invalidates the cache entry Authenticate populates for the
// revoked session, using the token hash the repository hands back — the
// gap noted in Authenticate's docs is closed now that it's available.
func (s *AccountService) RevokeSession(ctx context.Context, id model.ID) error {
	tokenHash, err := s.repo.RevokeSession(ctx, id)
	if err != nil {
		return err
	}

	if tokenHash != "" {
		s.cache.Del(ctx, model.AccountSessionKey(tokenHash))
	}

	return nil
}

// RevokeAllSessions invalidates the cache entry for every session it
// revoked; see RevokeSession.
func (s *AccountService) RevokeAllSessions(ctx context.Context, accountID model.ID) error {
	tokenHashes, err := s.repo.RevokeAllSessions(ctx, accountID)
	if err != nil {
		return err
	}

	for _, tokenHash := range tokenHashes {
		s.cache.Del(ctx, model.AccountSessionKey(tokenHash))
	}

	return nil
}

// IsMemberOfOrganization is cached, same reasoning as Authenticate's
// account/session lookups: adapter/http's auth middleware calls this on
// every request scoped to an organization. AddOrganization/RemoveOrganization
// invalidate the entry this populates.
func (s *AccountService) IsMemberOfOrganization(ctx context.Context, accountID, organizationID model.ID) (bool, error) {
	return utils.FetchThrough(ctx, s.cache, model.AccountOrganizationMemberKey(accountID, organizationID), func() (bool, error) {
		return s.repo.IsMemberOfOrganization(ctx, accountID, organizationID)
	})
}

func (s *AccountService) ListFromOrganization(ctx context.Context, organizationID model.ID) (model.List[model.Account], error) {
	return s.repo.ListFromOrganization(ctx, organizationID)
}

// AddOrganization invalidates the cache entry IsMemberOfOrganization
// populates, so the new membership is visible on the very next check
// instead of waiting out the cache's expiration.
func (s *AccountService) AddOrganization(ctx context.Context, accountID, organizationID model.ID) error {
	if err := s.repo.AddOrganization(ctx, accountID, organizationID); err != nil {
		return err
	}

	s.cache.Del(ctx, model.AccountOrganizationMemberKey(accountID, organizationID))

	return nil
}

// RemoveOrganization invalidates the cache entry IsMemberOfOrganization
// populates; see AddOrganization.
func (s *AccountService) RemoveOrganization(ctx context.Context, accountID, organizationID model.ID) error {
	if err := s.repo.RemoveOrganization(ctx, accountID, organizationID); err != nil {
		return err
	}

	s.cache.Del(ctx, model.AccountOrganizationMemberKey(accountID, organizationID))

	return nil
}

func (s *AccountService) ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Account], error) {
	return s.repo.ListFromTeam(ctx, teamID)
}

func (s *AccountService) AddTeam(ctx context.Context, accountID, teamID model.ID) error {
	return s.repo.AddTeam(ctx, accountID, teamID)
}

func (s *AccountService) RemoveTeam(ctx context.Context, accountID, teamID model.ID) error {
	return s.repo.RemoveTeam(ctx, accountID, teamID)
}

// newSessionToken generates a random session token (returned to the
// client) and its SHA-256 hash (what's actually persisted). Unlike
// passwords, session tokens are already high-entropy, so a fast hash is
// appropriate here — bcrypt is for low-entropy user-chosen secrets.
func newSessionToken() (token, tokenHash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token = hex.EncodeToString(raw)
	return token, hashSessionToken(token), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

var _ port.AccountService = (*AccountService)(nil)
