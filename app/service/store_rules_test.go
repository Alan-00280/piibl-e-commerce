package service

import (
	"testing"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

func TestCreateStoreValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})

	tests := []struct {
		name    string
		req     model.CreateStoresReq
		wantErr string
	}{
		{
			name: "valid store",
			req:  model.CreateStoresReq{Name: "Toko Sejahtera 99", Description: "Produk kebutuhan sehari-hari"},
		},
		{
			name:    "name is required",
			req:     model.CreateStoresReq{Description: "Produk kebutuhan sehari-hari"},
			wantErr: "name",
		},
		{
			name:    "name is too short",
			req:     model.CreateStoresReq{Name: "AB", Description: "Produk kebutuhan sehari-hari"},
			wantErr: "name",
		},
		{
			name:    "name cannot contain punctuation",
			req:     model.CreateStoresReq{Name: "Toko!", Description: "Produk kebutuhan sehari-hari"},
			wantErr: "name",
		},
		{
			name:    "name is too long",
			req:     model.CreateStoresReq{Name: "Toko Sejahtera dengan Nama yang Lebih dari Empat Puluh Lima Karakter", Description: "Produk kebutuhan sehari-hari"},
			wantErr: "name",
		},
		{
			name:    "description is required",
			req:     model.CreateStoresReq{Name: "Toko Sejahtera"},
			wantErr: "description",
		},
		{
			name:    "description is too short",
			req:     model.CreateStoresReq{Name: "Toko Sejahtera", Description: "AB"},
			wantErr: "description",
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

func TestPatchStoreValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	inactive := false
	validName := "Toko Sejahtera 99"
	invalidName := "Toko!"
	invalidDescription := "AB"

	tests := []struct {
		name    string
		req     model.PatchStoresReq
		wantErr string
	}{
		{name: "empty patch is valid for field validation"},
		{
			name: "inactive status can be explicitly false",
			req:  model.PatchStoresReq{IsActive: &inactive},
		},
		{
			name: "valid store name can contain spaces",
			req:  model.PatchStoresReq{Name: &validName},
		},
		{
			name:    "invalid store name",
			req:     model.PatchStoresReq{Name: &invalidName},
			wantErr: "name",
		},
		{
			name:    "invalid description",
			req:     model.PatchStoresReq{Description: &invalidDescription},
			wantErr: "description",
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
