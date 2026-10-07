package service

import "github.com/Alan-00280/piibl-e-commerce.git/app/model"

func isRoleAdmin(role model.Role) bool {
	return role == model.RoleAdmin
}
