package service

import (
	"context"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
)

func getTenantStores(tenantID int, storeRepo repository.StoreRepository, ctx context.Context) (model.Stores, error) {
	store, err := storeRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return model.Stores{}, translateErr(err, "store")
	}

	return store, nil
}

func checkVariantAdjustmentPrice(defaultPrice int64, adjustmentPrices []int64) string {
	var totalAdjustment int64
	totalAdjustment = 0

	for _, adjustmentPrice := range adjustmentPrices {
		totalAdjustment += adjustmentPrice
	}

	newPrice := totalAdjustment + defaultPrice
	if newPrice < 100 {
		return "Penyesuaian harga varian tidak boleh mengubah harga dasar hingga di bawah 100"
	}

	return ""
}

func IsEmptyPatchProduct(req model.ProductPatchReq) bool {
	return req.Name == nil && req.Description == nil && req.Status == nil && req.CategoryID == nil
}

func ApplyPatchProduct(current model.Product, req model.ProductPatchReq) model.Product {
	if req.Name != nil {
		current.Name = *req.Name
	}

	if req.Description != nil {
		current.Description = *req.Description
	}

	if req.Status != nil {
		current.Status = *req.Status
	}

	if req.CategoryID != nil {
		current.CategoryID = *req.CategoryID
	}

	return current
}
