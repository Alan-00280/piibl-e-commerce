package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound  = errors.New("data not found")
	ErrDuplicate = errors.New("data already exists")
	ErrInactive  = errors.New("data is being inactive")
	ErrVariant   = errors.New("varian produk tidak sah")
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
