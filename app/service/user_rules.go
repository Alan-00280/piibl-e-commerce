package service

import (
	"strings"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
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
