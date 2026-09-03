package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

var accountColumns = []string{"id", "name", "email", "phone_number", "type", "blocked", "created_at", "updated_at"}

// AccountRepository implements port.AccountRepository against Postgres.
type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account model.Account, passwordHash string) (model.Account, error) {
	query, args, err := psql.Insert("accounts").
		Columns("name", "email", "phone_number", "password_hash", "type", "blocked").
		Values(account.Name, account.Email, account.PhoneNumber, passwordHash, string(account.Type), account.Blocked).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Account{}, fmt.Errorf("postgres: failed to build create account query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Account{}, model.NewError(model.ErrCodeAccountEmailAlreadyExists, err)
		}
		return model.Account{}, fmt.Errorf("postgres: failed to create account: %w", err)
	}

	account.ID = model.ID(id)
	return account, nil
}

func (r *AccountRepository) Get(ctx context.Context, id model.ID) (model.Account, error) {
	query, args, err := psql.Select(accountColumns...).
		From("accounts").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Account{}, fmt.Errorf("postgres: failed to build get account query: %w", err)
	}

	account, err := scanAccount(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Account{}, model.NewError(model.ErrCodeAccountNotFound, err)
		}
		return model.Account{}, fmt.Errorf("postgres: failed to get account: %w", err)
	}

	return account, nil
}

// List ignores filters for now: model.AccountFilters carries no fields yet.
func (r *AccountRepository) List(ctx context.Context, filters model.AccountFilters) (model.List[model.Account], error) {
	query, args, err := psql.Select(accountColumns...).
		From("accounts").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to build list accounts query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to list accounts: %w", err)
	}
	defer rows.Close()

	return scanAccountList(rows)
}

func (r *AccountRepository) Update(ctx context.Context, account model.Account) (model.Account, error) {
	query, args, err := psql.Update("accounts").
		Set("name", account.Name).
		Set("email", account.Email).
		Set("phone_number", account.PhoneNumber).
		Set("type", string(account.Type)).
		Set("blocked", account.Blocked).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(account.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Account{}, fmt.Errorf("postgres: failed to build update account query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Account{}, model.NewError(model.ErrCodeAccountNotFound, err)
		}
		if isUniqueViolation(err) {
			return model.Account{}, model.NewError(model.ErrCodeAccountEmailAlreadyExists, err)
		}
		return model.Account{}, fmt.Errorf("postgres: failed to update account: %w", err)
	}

	return account, nil
}

func (r *AccountRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("accounts").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete account query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		if foreignKeyViolationConstraint(err) == "fk_account_organizations_account" {
			return model.NewError(model.ErrCodeAccountHasOrganizations, err)
		}
		return fmt.Errorf("postgres: failed to delete account: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeAccountNotFound, nil)
	}

	return nil
}

func (r *AccountRepository) GetPassword(ctx context.Context, id model.ID) (string, error) {
	query, args, err := psql.Select("password_hash").
		From("accounts").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("postgres: failed to build get account password query: %w", err)
	}

	var passwordHash string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.NewError(model.ErrCodeAccountNotFound, err)
		}
		return "", fmt.Errorf("postgres: failed to get account password: %w", err)
	}

	return passwordHash, nil
}

func (r *AccountRepository) SetPassword(ctx context.Context, id model.ID, passwordHash string) error {
	query, args, err := psql.Update("accounts").
		Set("password_hash", passwordHash).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build set account password query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to set account password: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeAccountNotFound, nil)
	}

	return nil
}

func (r *AccountRepository) GetByEmailWithPassword(ctx context.Context, email string) (model.Account, string, error) {
	query, args, err := psql.Select("id", "name", "email", "phone_number", "password_hash", "type", "blocked", "created_at", "updated_at").
		From("accounts").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return model.Account{}, "", fmt.Errorf("postgres: failed to build get account by email query: %w", err)
	}

	var (
		a            model.Account
		id           string
		accType      string
		passwordHash string
	)

	err = r.db.Pool.QueryRow(ctx, query, args...).
		Scan(&id, &a.Name, &a.Email, &a.PhoneNumber, &passwordHash, &accType, &a.Blocked, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Account{}, "", model.NewError(model.ErrCodeAccountNotFound, err)
		}
		return model.Account{}, "", fmt.Errorf("postgres: failed to get account by email: %w", err)
	}

	a.ID = model.ID(id)
	a.Type = model.AccountType(accType)
	return a, passwordHash, nil
}

func (r *AccountRepository) CreateSession(ctx context.Context, session model.AccountSession) (model.AccountSession, error) {
	query, args, err := psql.Insert("account_sessions").
		Columns("account_id", "token_hash", "user_agent", "ip_address", "expires_at").
		Values(string(session.AccountID), session.TokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.AccountSession{}, fmt.Errorf("postgres: failed to build create account session query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		if foreignKeyViolationConstraint(err) == "fk_account_sessions_account" {
			return model.AccountSession{}, model.NewError(model.ErrCodeAccountNotFound, err)
		}
		return model.AccountSession{}, fmt.Errorf("postgres: failed to create account session: %w", err)
	}

	session.ID = model.ID(id)
	return session, nil
}

var accountSessionColumns = []string{"id", "account_id", "token_hash", "user_agent", "ip_address", "expires_at", "revoked_at", "created_at", "updated_at"}

func (r *AccountRepository) GetSession(ctx context.Context, tokenHash string) (model.AccountSession, error) {
	query, args, err := psql.Select(accountSessionColumns...).
		From("account_sessions").
		Where(sq.Eq{"token_hash": tokenHash}).
		ToSql()
	if err != nil {
		return model.AccountSession{}, fmt.Errorf("postgres: failed to build get account session query: %w", err)
	}

	session, err := scanAccountSession(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AccountSession{}, model.NewError(model.ErrCodeAccountSessionNotFound, err)
		}
		return model.AccountSession{}, fmt.Errorf("postgres: failed to get account session: %w", err)
	}

	return session, nil
}

func (r *AccountRepository) ListSessions(ctx context.Context, accountID model.ID) (model.List[model.AccountSession], error) {
	query, args, err := psql.Select(accountSessionColumns...).
		From("account_sessions").
		Where(sq.Eq{"account_id": string(accountID)}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.AccountSession]{}, fmt.Errorf("postgres: failed to build list account sessions query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.AccountSession]{}, fmt.Errorf("postgres: failed to list account sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]model.AccountSession, 0)
	for rows.Next() {
		session, err := scanAccountSession(rows)
		if err != nil {
			return model.List[model.AccountSession]{}, fmt.Errorf("postgres: failed to scan account session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.AccountSession]{}, fmt.Errorf("postgres: failed to list account sessions: %w", err)
	}

	return model.List[model.AccountSession]{Items: sessions, Total: len(sessions)}, nil
}

// accountSessionExists is the shared idempotency check used by
// RevokeSession/RevokeSessionByToken: after an update-if-not-already-revoked
// affects 0 rows, this distinguishes "doesn't exist" from "already revoked"
// (both column and value are always internal, hardcoded constants — never
// user input).
func (r *AccountRepository) accountSessionExists(ctx context.Context, column, value string) (bool, error) {
	query, args, err := psql.Select("1").
		From("account_sessions").
		Where(sq.Eq{column: value}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("postgres: failed to build account session existence query: %w", err)
	}

	var exists bool
	if err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("postgres: failed to check account session existence: %w", err)
	}
	return exists, nil
}

// RevokeSession is idempotent: revoking an already-revoked session is a
// no-op (it does not overwrite the original RevokedAt), and only a
// nonexistent session ID is an error. The returned tokenHash is the
// revoked session's — "" if it was already revoked (nothing new to
// invalidate) — so callers can invalidate whatever they cache it under
// (see AccountService.Authenticate).
func (r *AccountRepository) RevokeSession(ctx context.Context, id model.ID) (string, error) {
	query, args, err := psql.Update("account_sessions").
		Set("revoked_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(id)}).
		Where("revoked_at IS NULL").
		Suffix("RETURNING token_hash").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("postgres: failed to build revoke account session query: %w", err)
	}

	var tokenHash string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&tokenHash)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("postgres: failed to revoke account session: %w", err)
		}

		// 0 rows updated: either the session doesn't exist, or it's
		// already revoked (idempotent no-op — already invalidated by
		// whichever call did the original revoke).
		exists, err := r.accountSessionExists(ctx, "id", string(id))
		if err != nil {
			return "", err
		}
		if !exists {
			return "", model.NewError(model.ErrCodeAccountSessionNotFound, nil)
		}
		return "", nil
	}

	return tokenHash, nil
}

// RevokeSessionByToken is RevokeSession keyed by token hash instead of
// session ID; same idempotency behavior.
func (r *AccountRepository) RevokeSessionByToken(ctx context.Context, tokenHash string) error {
	query, args, err := psql.Update("account_sessions").
		Set("revoked_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"token_hash": tokenHash}).
		Where("revoked_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build revoke account session query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to revoke account session: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	exists, err := r.accountSessionExists(ctx, "token_hash", tokenHash)
	if err != nil {
		return err
	}
	if !exists {
		return model.NewError(model.ErrCodeAccountSessionNotFound, nil)
	}

	return nil
}

// RevokeAllSessions revokes every active session for accountID (e.g. "log
// out everywhere", or on password change). Revoking zero sessions (none
// active) is not an error. The returned tokenHashes are the revoked
// sessions', so callers can invalidate whatever they cache them under
// (see AccountService.Authenticate).
func (r *AccountRepository) RevokeAllSessions(ctx context.Context, accountID model.ID) ([]string, error) {
	query, args, err := psql.Update("account_sessions").
		Set("revoked_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"account_id": string(accountID)}).
		Where("revoked_at IS NULL").
		Suffix("RETURNING token_hash").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to build revoke account sessions query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to revoke account sessions: %w", err)
	}
	defer rows.Close()

	tokenHashes := make([]string, 0)
	for rows.Next() {
		var tokenHash string
		if err := rows.Scan(&tokenHash); err != nil {
			return nil, fmt.Errorf("postgres: failed to scan revoked account session: %w", err)
		}
		tokenHashes = append(tokenHashes, tokenHash)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: failed to revoke account sessions: %w", err)
	}

	return tokenHashes, nil
}

func (r *AccountRepository) ListFromOrganization(ctx context.Context, organizationID model.ID) (model.List[model.Account], error) {
	query, args, err := psql.Select("a.id", "a.name", "a.email", "a.phone_number", "a.type", "a.blocked", "a.created_at", "a.updated_at").
		From("accounts a").
		Join("account_organizations ao ON ao.account_id = a.id").
		Where(sq.Eq{"ao.organization_id": string(organizationID)}).
		OrderBy("a.created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to build list accounts from organization query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to list accounts from organization: %w", err)
	}
	defer rows.Close()

	return scanAccountList(rows)
}

// AddOrganization is idempotent: adding an account to an organization it
// already belongs to is a no-op, not an error.
func (r *AccountRepository) AddOrganization(ctx context.Context, accountID, organizationID model.ID) error {
	query, args, err := psql.Insert("account_organizations").
		Columns("account_id", "organization_id").
		Values(string(accountID), string(organizationID)).
		Suffix("ON CONFLICT (account_id, organization_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build add account to organization query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to add account to organization: %w", err)
	}
	return nil
}

// RemoveOrganization is idempotent: removing a membership that doesn't
// exist is a no-op, not an error.
func (r *AccountRepository) RemoveOrganization(ctx context.Context, accountID, organizationID model.ID) error {
	query, args, err := psql.Delete("account_organizations").
		Where(sq.Eq{"account_id": string(accountID), "organization_id": string(organizationID)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build remove account from organization query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to remove account from organization: %w", err)
	}
	return nil
}

func (r *AccountRepository) ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Account], error) {
	query, args, err := psql.Select("a.id", "a.name", "a.email", "a.phone_number", "a.type", "a.blocked", "a.created_at", "a.updated_at").
		From("accounts a").
		Join("account_teams act ON act.account_id = a.id").
		Where(sq.Eq{"act.team_id": string(teamID)}).
		OrderBy("a.created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to build list accounts from team query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to list accounts from team: %w", err)
	}
	defer rows.Close()

	return scanAccountList(rows)
}

// AddTeam is idempotent: adding an account to a team it already belongs to
// is a no-op, not an error.
func (r *AccountRepository) AddTeam(ctx context.Context, accountID, teamID model.ID) error {
	query, args, err := psql.Insert("account_teams").
		Columns("account_id", "team_id").
		Values(string(accountID), string(teamID)).
		Suffix("ON CONFLICT (account_id, team_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build add account to team query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to add account to team: %w", err)
	}
	return nil
}

// RemoveTeam is idempotent: removing a membership that doesn't exist is a
// no-op, not an error.
func (r *AccountRepository) RemoveTeam(ctx context.Context, accountID, teamID model.ID) error {
	query, args, err := psql.Delete("account_teams").
		Where(sq.Eq{"account_id": string(accountID), "team_id": string(teamID)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build remove account from team query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to remove account from team: %w", err)
	}
	return nil
}

// scannableRow is satisfied by both pgx.Row (QueryRow) and pgx.Rows
// (Query, one row at a time), letting scanAccount back both.
type scannableRow interface {
	Scan(dest ...any) error
}

func scanAccount(row scannableRow) (model.Account, error) {
	var (
		a       model.Account
		id      string
		accType string
	)

	err := row.Scan(&id, &a.Name, &a.Email, &a.PhoneNumber, &accType, &a.Blocked, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return model.Account{}, err
	}

	a.ID = model.ID(id)
	a.Type = model.AccountType(accType)
	return a, nil
}

func scanAccountList(rows pgx.Rows) (model.List[model.Account], error) {
	accounts := make([]model.Account, 0)
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return model.List[model.Account]{}, fmt.Errorf("postgres: failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Account]{}, fmt.Errorf("postgres: failed to list accounts: %w", err)
	}

	return model.List[model.Account]{Items: accounts, Total: len(accounts)}, nil
}

func scanAccountSession(row scannableRow) (model.AccountSession, error) {
	var (
		s         model.AccountSession
		id        string
		accountID string
	)

	err := row.Scan(&id, &accountID, &s.TokenHash, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return model.AccountSession{}, err
	}

	s.ID = model.ID(id)
	s.AccountID = model.ID(accountID)
	return s, nil
}

var _ port.AccountRepository = (*AccountRepository)(nil)
