package service

import (
	"errors"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
)

// func CanAccess(model.AuthUser, targetID int, *helper.PermissionSet, anyPermission string) bool
// memeriksa id user terautentikasi dengan target id
// diijinkan ke target id sama
// diijinkan ke permission *:any
func CanAccess(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}

	return perms.Can(current.Role, anyPermission)
}

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

func checkGetCurrentUser(c *fiber.Ctx) (model.AuthUser, error) {
	currentUser, ok := helper.CurentUser(c)
	if !ok {
		return model.AuthUser{}, helper.Unauthorized("anda tidak terautentikasi")
	}

	return currentUser, nil
}
