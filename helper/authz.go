package helper

import (
	"sort"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
)

// Menyimpan data role permission menggunakan struct
// struct bernama PermissionSet
// berisi byRole: map[string]map[string]struct{}
//
//	contoh bentuk:
//	byRole = {
//	    'admin' = {
//	          'user:list' = {},
//	          'user:read:any' = {}
//	     },
//	}
type PermissionSet struct {
	byRole map[string]map[string]struct{}
}

// func NewPermissionSet(raw map[string][]string) *PermissionSet
// untuk membuat struct PermissionSet dari raw result query database
func NewPermissionSet(raw map[string][]string) *PermissionSet {
	byRole := make(map[string]map[string]struct{})

	for role, permissions := range raw {
		set := make(map[string]struct{})

		for _, permission := range permissions {
			set[permission] = struct{}{}
		}

		byRole[role] = set
	}

	return &PermissionSet{
		byRole: byRole,
	}
}

// Method:
// p == nil --> must return
// Can(role, permission string) bool
// Memberikan nilai boolean
// Jika role punya permission --> true dan sebaliknya
func (p *PermissionSet) Can(role model.Role, permission string) bool {
	if p == nil {
		return false
	}

	permissions, ok := p.byRole[string(role)]
	if !ok {
		return false
	}

	_, ok = permissions[permission]
	if !ok {
		return false
	}

	return true
}

// PermissionsOf(role string) []string
// Mengembalikan array string berisi permissions suatu role
func (p *PermissionSet) PermissionsOf(role model.Role) []string {
	result := []string{}

	if p == nil {
		return []string{}
	}

	permissions, ok := p.byRole[string(role)]
	if !ok {
		return []string{}
	}

	for permission := range permissions {
		result = append(result, permission)
	}

	sort.Strings(result)
	return result
}

// KnownRoles() []string
// Mengembalikan semua role yang ada
func (p *PermissionSet) KnownRoles() []string {
	result := []string{}

	if p == nil {
		return []string{}
	}

	for role := range p.byRole {
		result = append(result, role)
	}

	sort.Strings(result)
	return result
}

// IsKnownRoles(role string) bool
// Mengembalikan boolean
// jika role ada --> true dan sebaliknya
func (p *PermissionSet) IsKnownRoles(role model.Role) bool {
	if p == nil {
		return false
	}

	_, exists := p.byRole[string(role)]
	return exists
}

// Mengembalikan model.Role
// dari string
// dan hanya string valid yang bisa
// ADMIN | CUSTOMER | TENANT
func RoleOf(role string) (model.Role, bool) {
	switch {
	case role == string(model.RoleAdmin):
		return model.RoleAdmin, true
	case role == string(model.RoleCustomer):
		return model.RoleCustomer, true
	case role == string(model.RoleTenant):
		return model.RoleTenant, true
	default:
		return model.Role(""), false
	}
}
