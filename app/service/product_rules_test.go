package service

import (
	"testing"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

func TestProductCreateValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	categoryID := 2
	invalidCategory := 0

	tests := []struct {
		name    string
		req     model.ProductCreateReq
		wantErr string
	}{
		{
			name: "valid product",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika 250", Description: "Kopi pilihan dalam kemasan 250 gram",
				BasePrice: 15000, Stock: 10, Status: model.ProductStatActive,
			},
		},
		{
			name: "zero stock is valid",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				BasePrice: 15000, Stock: 0, Status: model.ProductStatActive,
			},
		},
		{
			name: "stock lower bound above zero is valid",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				BasePrice: 15000, Stock: 1, Status: model.ProductStatActive,
			},
		},
		{
			name: "stock upper bound is valid",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				BasePrice: 15000, Stock: 100000, Status: model.ProductStatActive,
			},
		},
		{
			name: "base price is required",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "base_price",
		},
		{
			name: "base price must be at least 100",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				BasePrice: 99, Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "base_price",
		},
		{
			name: "base price cannot exceed upper bound",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				BasePrice: 1000000001, Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "base_price",
		},
		{
			name: "store ID must be positive",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "store_id",
		},
		{
			name: "category ID must be positive",
			req: model.ProductCreateReq{
				CategoryID: &invalidCategory, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "category_id",
		},
		{
			name: "name must be alphanumeric with spaces",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika!", Description: "Kopi pilihan dalam kemasan 250 gram",
				Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "name",
		},
		{
			name: "description is required",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Stock: 10, Status: model.ProductStatActive,
			},
			wantErr: "description",
		},
		{
			name: "stock cannot exceed upper bound",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				Stock: 100001, Status: model.ProductStatActive,
			},
			wantErr: "stock",
		},
		{
			name: "status is required",
			req: model.ProductCreateReq{
				CategoryID: &categoryID, Name: "Kopi Arabika", Description: "Kopi pilihan dalam kemasan 250 gram",
				Stock: 10,
			},
			wantErr: "status",
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

func TestProductPatchValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})
	invalidName := "Kopi!"
	invalidCategoryID := 0
	inactive := model.ProductStatInactive

	tests := []struct {
		name    string
		req     model.ProductPatchReq
		wantErr string
	}{
		{name: "empty patch is valid for field validation"},
		{
			name: "status may be explicitly set to inactive",
			req:  model.ProductPatchReq{Status: &inactive},
		},
		{
			name:    "category ID must be positive",
			req:     model.ProductPatchReq{CategoryID: &invalidCategoryID},
			wantErr: "category_id",
		},
		{
			name:    "name cannot contain punctuation",
			req:     model.ProductPatchReq{Name: &invalidName},
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

func TestProductStatusAndStockValidation(t *testing.T) {
	validator := helper.NewValidator(&helper.PasswordCommonSet{PasswordSet: map[string]struct{}{}})

	tests := []struct {
		name    string
		req     any
		wantErr string
	}{
		{
			name:    "deactivation status is required",
			req:     model.ProductDeactivateReq{},
			wantErr: "status",
		},
		{
			name: "inactive deactivation status is valid",
			req:  model.ProductDeactivateReq{Status: model.ProductStatInactive},
		},
		{
			name: "zero stock update is valid",
			req:  model.ProductUpdateStockReq{Stock: 0},
		},
		{
			name: "positive stock update is valid",
			req:  model.ProductUpdateStockReq{Stock: 1},
		},
		{
			name:    "stock update cannot exceed upper bound",
			req:     model.ProductUpdateStockReq{Stock: 100001},
			wantErr: "stock",
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

func TestCheckVariantAdjustmentPrice(t *testing.T) {
	const minimumPrice int64 = 100
	const belowMinimumMessage = "Penyesuaian harga varian tidak boleh mengubah harga dasar hingga di bawah 100"

	tests := []struct {
		name             string
		defaultPrice     int64
		adjustmentPrices []int64
		want             string
	}{
		{
			name:         "base price at minimum with no adjustments",
			defaultPrice: minimumPrice,
		},
		{
			name:             "adjustment total reaches minimum",
			defaultPrice:     150,
			adjustmentPrices: []int64{-25, -25},
		},
		{
			name:             "positive adjustment",
			defaultPrice:     100,
			adjustmentPrices: []int64{50},
		},
		{
			name:             "adjustments reduce price below minimum",
			defaultPrice:     150,
			adjustmentPrices: []int64{-25, -26},
			want:             belowMinimumMessage,
		},
		{
			name:             "base price already below minimum",
			defaultPrice:     99,
			adjustmentPrices: []int64{},
			want:             belowMinimumMessage,
		},
		{
			name:             "combined positive and negative adjustments",
			defaultPrice:     120,
			adjustmentPrices: []int64{30, -51},
			want:             belowMinimumMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkVariantAdjustmentPrice(tt.defaultPrice, tt.adjustmentPrices); got != tt.want {
				t.Errorf("checkVariantAdjustmentPrice(%d, %v) = %q, want %q",
					tt.defaultPrice, tt.adjustmentPrices, got, tt.want)
			}
		})
	}
}
