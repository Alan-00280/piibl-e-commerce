package service

import (
	"reflect"
	"testing"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

var passwordCommonSet, _ = helper.NewPasswordCommonSet("../../files/common_password.txt")

func TestApplyPatchUser(t *testing.T) {
	current := model.User{
		ID:       7,
		Username: "olduser",
		Email:    "old@example.com",
		Role:     model.RoleCustomer,
		IsActive: true,
	}
	newUsername := "  newuser  "
	newEmail := "new@example.com"
	deactivate := false

	got := ApplyPatchUser(current, model.PatchUserRequest{
		Username: &newUsername,
		Email:    &newEmail,
		IsActive: &deactivate,
	})

	want := current
	want.Username = "newuser"
	want.Email = newEmail
	want.IsActive = false
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ApplyPatchUser() = %#v, want %#v", got, want)
	}
}

func TestApplyPatchUserLeavesOmittedFieldsUnchanged(t *testing.T) {
	current := model.User{
		ID:       7,
		Username: "existing",
		Email:    "existing@example.com",
		Role:     model.RoleTenant,
		IsActive: true,
	}

	got := ApplyPatchUser(current, model.PatchUserRequest{})
	if !reflect.DeepEqual(got, current) {
		t.Errorf("ApplyPatchUser() = %#v, want unchanged user %#v", got, current)
	}
}

func TestIsEmptyPatchUser(t *testing.T) {
	username := ""
	email := ""
	active := false
	tests := []struct {
		name string
		req  model.PatchUserRequest
		want bool
	}{
		{name: "empty patch", req: model.PatchUserRequest{}, want: true},
		{name: "username provided", req: model.PatchUserRequest{Username: &username}},
		{name: "email provided", req: model.PatchUserRequest{Email: &email}},
		{name: "active status provided as false", req: model.PatchUserRequest{IsActive: &active}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyPatchUser(tt.req); got != tt.want {
				t.Errorf("IsEmptyPatchUser() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestCreateUserValidation(t *testing.T) {
	validator := helper.NewValidator(passwordCommonSet)

	tests := []struct {
		name    string
		req     model.CreateUserReq
		wantErr string
	}{
		{
			name: "valid user",
			req:  model.CreateUserReq{Username: "user123", Email: "user@example.com", Password: "strongpass123"},
		},
		{
			name:    "invalid username",
			req:     model.CreateUserReq{Username: "ab", Email: "user@example.com", Password: "strongpass123"},
			wantErr: "username",
		},
		{
			name:    "invalid email",
			req:     model.CreateUserReq{Username: "user123", Email: "not-an-email", Password: "strongpass123"},
			wantErr: "email",
		},
		{
			name:    "weak common password",
			req:     model.CreateUserReq{Username: "user123", Email: "user@example.com", Password: "password123"},
			wantErr: "password",
		},
		{
			name:    "password without a digit",
			req:     model.CreateUserReq{Username: "user123", Email: "user@example.com", Password: "weakpassword"},
			wantErr: "password",
		},
		{
			name:    "password without a letter",
			req:     model.CreateUserReq{Username: "user123", Email: "user@example.com", Password: "121212456"},
			wantErr: "password",
		},
		{
			name:    "password is to short",
			req:     model.CreateUserReq{Username: "user123", Email: "user@example.com", Password: "123"},
			wantErr: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.ValidateStruct(tt.req, *validator)
			if tt.wantErr == "" {
				if len(got) != 0 {
					t.Errorf("ValidateStruct() errors = %v, want no errors", got)
				}
				return
			}

			if _, ok := got[tt.wantErr]; !ok {
				t.Errorf("ValidateStruct() errors = %v, want an error for %q", got, tt.wantErr)
			}
		})
	}
}

func TestPatchUserValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	active := false
	invalidUsername := "ab"
	invalidEmail := "not-an-email"

	tests := []struct {
		name    string
		req     model.PatchUserRequest
		wantErr string
	}{
		{name: "no fields provided"},
		{
			name: "active status explicitly false",
			req:  model.PatchUserRequest{IsActive: &active},
		},
		{
			name:    "invalid username",
			req:     model.PatchUserRequest{Username: &invalidUsername},
			wantErr: "username",
		},
		{
			name:    "invalid email",
			req:     model.PatchUserRequest{Email: &invalidEmail},
			wantErr: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.ValidateStruct(tt.req, *validator)
			if tt.wantErr == "" {
				if len(got) != 0 {
					t.Errorf("ValidateStruct() errors = %v, want no errors", got)
				}
				return
			}

			if _, ok := got[tt.wantErr]; !ok {
				t.Errorf("ValidateStruct() errors = %v, want an error for %q", got, tt.wantErr)
			}
		})
	}
}
