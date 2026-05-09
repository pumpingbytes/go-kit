package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgErrCodeForeignKeyViolation = "23503"
	pgErrCodeUniqueViolation     = "23505"
)

type Classifier struct{}

func (c Classifier) IsNoRowsError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (c Classifier) IsForeignKeyViolationError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgErrCodeForeignKeyViolation
}

func (c Classifier) IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation
}

func New() *Classifier {
	return &Classifier{}
}
