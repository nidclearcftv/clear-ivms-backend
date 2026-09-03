package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres SQLSTATE codes this adapter cares about.
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgErrCodeUniqueViolation     = "23505"
	pgErrCodeForeignKeyViolation = "23503"
)

// asPgError unwraps err into a *pgconn.PgError, if it is (or wraps) one.
func asPgError(err error) (*pgconn.PgError, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr, true
	}
	return nil, false
}

func isUniqueViolation(err error) bool {
	pgErr, ok := asPgError(err)
	return ok && pgErr.Code == pgErrCodeUniqueViolation
}

// foreignKeyViolationConstraint returns the violated constraint's name if
// err is a foreign key violation, and "" otherwise.
func foreignKeyViolationConstraint(err error) string {
	pgErr, ok := asPgError(err)
	if !ok || pgErr.Code != pgErrCodeForeignKeyViolation {
		return ""
	}
	return pgErr.ConstraintName
}
