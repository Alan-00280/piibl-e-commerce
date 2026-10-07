package repository

import "errors"

var (
	ErrNotFound  = errors.New("data not found")
	ErrDuplicate = errors.New("data already exists")
	ErrVariant   = errors.New("varian produk tidak sah")
)
