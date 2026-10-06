package service

import (
	"testing"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

func TestProductVariantSpaceCreateValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	mandatory := true
	optional := false

	tests := []struct {
		name    string
		req     model.ProductVariantSpaceCreateReq
		wantErr string
	}{
		{
			name: "valid mandatory variant space",
			req:  model.ProductVariantSpaceCreateReq{ProductID: 1, Name: "Ukuran Produk", IsMandatory: &mandatory},
		},
		{
			name: "valid optional variant space",
			req:  model.ProductVariantSpaceCreateReq{ProductID: 1, Name: "Ukuran Produk", IsMandatory: &optional},
		},
		{
			name:    "product ID must be positive",
			req:     model.ProductVariantSpaceCreateReq{Name: "Ukuran Produk", IsMandatory: &mandatory},
			wantErr: "product_id",
		},
		{
			name:    "name is required",
			req:     model.ProductVariantSpaceCreateReq{ProductID: 1, IsMandatory: &mandatory},
			wantErr: "name",
		},
		{
			name:    "name must be at least three characters",
			req:     model.ProductVariantSpaceCreateReq{ProductID: 1, Name: "AB", IsMandatory: &mandatory},
			wantErr: "name",
		},
		{
			name:    "name cannot contain punctuation",
			req:     model.ProductVariantSpaceCreateReq{ProductID: 1, Name: "Ukuran!", IsMandatory: &mandatory},
			wantErr: "name",
		},
		{
			name:    "mandatory flag is required",
			req:     model.ProductVariantSpaceCreateReq{ProductID: 1, Name: "Ukuran Produk"},
			wantErr: "is_mandatory",
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

func TestProductVariantSpacePatchValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	optional := false
	validName := "Ukuran Produk"
	invalidName := "AB"

	tests := []struct {
		name    string
		req     model.ProductVariantSpacePatchReq
		wantErr string
	}{
		{name: "empty patch is valid for field validation"},
		{
			name: "optional flag can explicitly be false",
			req:  model.ProductVariantSpacePatchReq{IsMandatory: &optional},
		},
		{
			name: "name can contain spaces",
			req:  model.ProductVariantSpacePatchReq{Name: &validName},
		},
		{
			name:    "name must be at least three characters",
			req:     model.ProductVariantSpacePatchReq{Name: &invalidName},
			wantErr: "name",
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

func TestProductVariantCreateValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	spaceID := 2
	zeroSpaceID := 0
	minAdjustment := int64(-1000000000)
	maxAdjustment := int64(100000000000)
	tooSmallAdjustment := int64(-1000000001)
	tooLargeAdjustment := int64(100000000001)

	tests := []struct {
		name    string
		req     model.ProductVariantCreateReq
		wantErr string
	}{
		{
			name: "valid variant without space or price adjustment",
			req:  model.ProductVariantCreateReq{ProductID: 1, ProductVariantItems: model.ProductVariantItem{Name: "Large"}},
		},
		{
			name: "valid variant with space and price adjustment",
			req: model.ProductVariantCreateReq{
				ProductID: 1, VariantSpaceID: &spaceID,
				ProductVariantItems: model.ProductVariantItem{Name: "Large", PriceAdjusment: &minAdjustment},
			},
		},
		{
			name: "maximum price adjustment is valid",
			req: model.ProductVariantCreateReq{
				ProductID:           1,
				ProductVariantItems: model.ProductVariantItem{Name: "Large", PriceAdjusment: &maxAdjustment},
			},
		},
		{
			name:    "product ID must be positive",
			req:     model.ProductVariantCreateReq{ProductVariantItems: model.ProductVariantItem{Name: "Large"}},
			wantErr: "product_id",
		},
		{
			name: "variant space ID must be positive when provided",
			req: model.ProductVariantCreateReq{
				ProductID: 1, VariantSpaceID: &zeroSpaceID,
				ProductVariantItems: model.ProductVariantItem{Name: "Large"},
			},
			wantErr: "space_id",
		},
		{
			name:    "variant name is required",
			req:     model.ProductVariantCreateReq{ProductID: 1},
			wantErr: "name",
		},
		{
			name: "variant name must be at least three characters",
			req: model.ProductVariantCreateReq{
				ProductID: 1, ProductVariantItems: model.ProductVariantItem{Name: "AB"},
			},
			wantErr: "name",
		},
		{
			name: "price adjustment cannot be below minimum",
			req: model.ProductVariantCreateReq{
				ProductID:           1,
				ProductVariantItems: model.ProductVariantItem{Name: "Large", PriceAdjusment: &tooSmallAdjustment},
			},
			wantErr: "price_adjustment",
		},
		{
			name: "price adjustment cannot exceed maximum",
			req: model.ProductVariantCreateReq{
				ProductID:           1,
				ProductVariantItems: model.ProductVariantItem{Name: "Large", PriceAdjusment: &tooLargeAdjustment},
			},
			wantErr: "price_adjustment",
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

func TestProductVariantPatchValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	validName := "Large Size"
	invalidName := "AB"
	minAdjustment := int64(-100000000000)
	maxAdjustment := int64(100000000000)
	tooSmallAdjustment := int64(-100000000001)
	tooLargeAdjustment := int64(100000000001)

	tests := []struct {
		name    string
		req     model.ProductVariantPatchReq
		wantErr string
	}{
		{name: "empty patch is valid for field validation"},
		{
			name: "name and price adjustment are optional",
			req:  model.ProductVariantPatchReq{Name: &validName, PriceAdjusment: &minAdjustment},
		},
		{
			name: "maximum price adjustment is valid",
			req:  model.ProductVariantPatchReq{PriceAdjusment: &maxAdjustment},
		},
		{
			name:    "name must be at least three characters",
			req:     model.ProductVariantPatchReq{Name: &invalidName},
			wantErr: "name",
		},
		{
			name:    "price adjustment cannot be below minimum",
			req:     model.ProductVariantPatchReq{PriceAdjusment: &tooSmallAdjustment},
			wantErr: "price",
		},
		{
			name:    "price adjustment cannot exceed maximum",
			req:     model.ProductVariantPatchReq{PriceAdjusment: &tooLargeAdjustment},
			wantErr: "price",
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
