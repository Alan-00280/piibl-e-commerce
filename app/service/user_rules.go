package service

import (
	"strings"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

func ApplyPatchUser(current model.User, req model.PatchUserRequest) model.User {
	if req.Username != nil {
		current.Username = strings.TrimSpace(*req.Username)
	}

	if req.Email != nil {
		current.Email = *req.Email
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current
}

func IsEmptyPatchUser(req model.PatchUserRequest) bool {
	return req.Username == nil && req.Email == nil && req.IsActive == nil
}

// func ValidateAssignRole(model.AuthUser, targetID int, model.AssignRoleRequest, *helper.PermissionSet) map[string]string
// memeriksa current id dengan target id
//
//	jika sama ditolak (mengembalikan map error)
//
// memeriksa validitas nama role
// trim space role
func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	req model.AssignRoleRequest,
	perms *helper.PermissionSet,
) map[string]string {
	errs := map[string]string{}

	if !perms.IsKnownRoles(req.Role) {
		errs["role"] = string(req.Role) + " tidak termasuk dalam role valid: " + strings.Join(perms.KnownRoles(), ", ")
	}

	if current.UserID == targetID {
		errs["role"] = "user tidak dapat mengubah role milik sendiri"
	}

	return errs
}
