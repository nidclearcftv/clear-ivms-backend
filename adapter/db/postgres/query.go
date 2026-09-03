package postgres

import sq "github.com/Masterminds/squirrel"

// psql is the shared query builder used for every query in this package —
// static and conditional alike — so there's one consistent pattern rather
// than a mix of builder calls and raw SQL strings. Configured for
// Postgres's $1, $2, ... placeholder style; squirrel defaults to "?"
// (MySQL/SQLite style).
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
