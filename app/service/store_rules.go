package service

import (
	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

func IsEmptyPatchStore(req model.PatchStoresReq) bool {
	return req.Name == nil && req.Description == nil && req.IsActive == nil
}

func CanAccessStore(
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

func ApplyPatchStore(current model.Stores, req model.PatchStoresReq) model.Stores {
	if req.Name != nil {
		current.Name = *req.Name
	}

	if req.Description != nil {
		current.Description = *req.Description
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current
}
