package service

import (
	"errors"

	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

func translateErr(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " can't be found")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("data already used")
	default:
		return helper.Internal(err)
	}
}

// make the total page even without decimal number
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}

	return (total + limit - 1) / limit
}
