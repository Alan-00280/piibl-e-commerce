package service

import (
	"context"
	"fmt"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
)

// ==============================
// VARIANTS PRODUK
// ==============================

// POST /variant-spaces
// roles: tenant
// perms:
// rules:
//  1. Produk milik sendiri
//  2. Harga harus dihitung agar tidak negatif
func (s *ProductService) CreateSVariant(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	productID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ProductVariantSpaceCreateReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, productID)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "variant:manage:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	priceError, err := s.validateProductVariantPrices(ctx, productID)
	if err != nil {
		return translateErr(err, "product variants")
	}
	if priceError != "" {
		return helper.BadRequest(priceError)
	}

	variantSpace, err := s.repo.CreateVariantSpace(ctx, model.ProductVariantSpace{
		ProductID:   productID,
		Name:        req.Name,
		IsMandatory: *req.IsMandatory,
	})
	if err != nil {
		return translateErr(err, "variant space")
	}

	return helper.Created(
		c,
		"variant space berhasil dibuat",
		variantSpace,
		fmt.Sprintf("/api/v1/products/%d/variant-spaces/%d", productID, variantSpace.ID),
	)
}

// PATCH /variant-spaces/:id
// roles: tenant
// perms
// rules:
//  1. Produk sendiri
//  2. Harga harus dihitung agar tidak negatif
func (s *ProductService) PatchSVariant(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	variantSpaceID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ProductVariantSpacePatchReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}
	if req.Name == nil && req.IsMandatory == nil {
		return helper.BadRequest("body kosong")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	variantSpace, err := s.repo.FindVariantSpaceByID(ctx, variantSpaceID)
	if err != nil {
		return translateErr(err, "variant space")
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, variantSpace.ProductID)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "variant:manage:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	priceError, err := s.validateProductVariantPrices(ctx, variantSpace.ProductID)
	if err != nil {
		return translateErr(err, "product variants")
	}
	if priceError != "" {
		return helper.BadRequest(priceError)
	}

	if req.Name != nil {
		variantSpace.Name = *req.Name
	}
	if req.IsMandatory != nil {
		variantSpace.IsMandatory = *req.IsMandatory
	}

	updated, err := s.repo.UpdateVariantSpace(ctx, variantSpace)
	if err != nil {
		return translateErr(err, "variant space")
	}

	return helper.Ok(c, "variant space berhasil diperbarui", updated)
}

// DELETE /variant-spaces/:id
// roles: tenant, admin
// perms:
//
//	tenant -> ownership check
//	admin  -> variant:manage:any
func (s *ProductService) DeleteSVaraint(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	variantSpaceID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	variantSpace, err := s.repo.FindVariantSpaceByID(ctx, variantSpaceID)
	if err != nil {
		return translateErr(err, "variant space")
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, variantSpace.ProductID)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "variant:manage:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	if err := s.repo.DeleteVariantSpace(ctx, variantSpaceID); err != nil {
		return translateErr(err, "variant space")
	}

	return helper.NoContent(c)
}

// POST /variants
// roles: tenant, admin
// perms:
//
//	tenant -> ownership check
//	admin  -> variant:manage:any
//
// rules:
//  1. Produk harus milik tenant
//  2. Variant yang dibuat adalah variant biasa
//  3. Tidak boleh membuat variant bernama _default
func (s *ProductService) CreateVariant(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	var req model.ProductVariantCreateReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}
	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}
	if len(req.ProductVariantItems) == 0 {
		return helper.BadRequest("minimal satu variant harus dibuat")
	}
	if req.VariantSpaceID == nil {
		return helper.BadRequest("variant harus memiliki variant space")
	}
	var validationErrors map[string]string
	for index, item := range req.ProductVariantItems {
		if item.Name == "_default" {
			return helper.BadRequest("nama _default hanya digunakan untuk harga dasar produk")
		}
		if errs := helper.ValidateStruct(item, *s.appValidator); len(errs) > 0 {
			if validationErrors == nil {
				validationErrors = make(map[string]string)
			}
			for field, message := range errs {
				validationErrors[fmt.Sprintf("product_variant_items[%d].%s", index, field)] = message
			}
		}
	}
	if len(validationErrors) > 0 {
		return helper.Validation(validationErrors)
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}
	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, req.ProductID)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "variant:manage:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	variantSpace, err := s.repo.FindVariantSpaceByID(ctx, *req.VariantSpaceID)
	if err != nil {
		return translateErr(err, "variant space")
	}
	if variantSpace.ProductID != req.ProductID {
		return helper.BadRequest("variant space bukan milik product ini")
	}

	variants := make([]model.ProductVariant, 0, len(req.ProductVariantItems))
	for _, item := range req.ProductVariantItems {
		variant := model.ProductVariant{
			ProductID:      req.ProductID,
			VariantSpaceID: req.VariantSpaceID,
			Name:           item.Name,
			PriceAdjusment: item.PriceAdjusment,
		}
		priceError, err := s.validateProductVariantPricesWithChange(ctx, req.ProductID, &variant, 0)
		if err != nil {
			return translateErr(err, "product variants")
		}
		if priceError != "" {
			return helper.BadRequest(priceError)
		}
		variants = append(variants, variant)
	}

	createdVariants := make([]model.ProductVariant, 0, len(variants))
	for _, variant := range variants {
		created, err := s.repo.CreateVariant(ctx, variant)
		if err != nil {
			return translateErr(err, "variant")
		}
		createdVariants = append(createdVariants, created)
	}

	return helper.Created(
		c,
		"variants berhasil dibuat",
		createdVariants,
		"/api/v1/variants",
	)
}

// PATCH /variants/:id
// roles: tenant, admin
// perms:
//
//	tenant -> ownership check
//	admin  -> variant:manage:any
//
// rules:
//  1. Produk harus milik tenant
//  2. Variant harus berada di product/space tersebut
//  3. _default tidak boleh diperlakukan sebagai variant biasa
func (s *ProductService) PatchVariant(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	variantID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ProductVariantPatchReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}
	if req.Name == nil && req.PriceAdjusment == nil {
		return helper.BadRequest("body kosong")
	}
	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	variant, err := s.repo.FindVariantByID(ctx, variantID)
	if err != nil {
		return translateErr(err, "variant")
	}
	if variant.Name == "_default" || variant.VariantSpaceID == nil {
		return helper.BadRequest("variant _default tidak dapat diubah sebagai variant biasa")
	}
	if req.Name != nil && *req.Name == "_default" {
		return helper.BadRequest("nama _default hanya digunakan untuk harga dasar produk")
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, variant.ProductID)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "variant:manage:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	if req.Name != nil {
		variant.Name = *req.Name
	}
	if req.PriceAdjusment != nil {
		variant.PriceAdjusment = req.PriceAdjusment
	}

	priceError, err := s.validateProductVariantPricesWithChange(ctx, variant.ProductID, &variant, variant.ID)
	if err != nil {
		return translateErr(err, "product variants")
	}
	if priceError != "" {
		return helper.BadRequest(priceError)
	}

	updated, err := s.repo.UpdateVariant(ctx, variant)
	if err != nil {
		return translateErr(err, "variant")
	}

	return helper.Ok(c, "variant berhasil diperbarui", updated)
}

// DELETE /variants/:id
// roles: tenant, admin
// perms:
//
//	tenant -> ownership check
//	admin  -> variant:manage:any
//
// rules:
//  1. Produk harus milik tenant
//  2. _default tidak boleh dihapus
//  3. Variant biasa boleh dihapus permanen
func (s *ProductService) DeleteVariant(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	variantID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	variant, err := s.repo.FindVariantByID(ctx, variantID)
	if err != nil {
		return translateErr(err, "variant")
	}
	if variant.Name == "_default" || variant.VariantSpaceID == nil {
		return helper.BadRequest("variant _default tidak dapat dihapus")
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, variant.ProductID)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "variant:manage:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	if err := s.repo.DeleteVariant(ctx, variantID); err != nil {
		return translateErr(err, "variant")
	}

	return helper.NoContent(c)
}

func (s *ProductService) validateProductVariantPrices(ctx context.Context, productID int) (string, error) {
	return s.validateProductVariantPricesWithChange(ctx, productID, nil, 0)
}

func (s *ProductService) validateProductVariantPricesWithChange(
	ctx context.Context,
	productID int,
	changedVariant *model.ProductVariant,
	excludedVariantID int,
) (string, error) {
	defaultVariant, err := s.repo.FindDefaultVariant(ctx, productID)
	if err != nil {
		return "", err
	}
	if defaultVariant.PriceAdjusment == nil {
		return "", fmt.Errorf("default variant for product %d has no base price", productID)
	}

	variants, err := s.repo.FindVariantsByProductID(ctx, productID)
	if err != nil {
		return "", err
	}

	minimumAdjustmentBySpace := make(map[int]int64)
	for _, variant := range variants {
		if variant.VariantSpaceID == nil || variant.Name == "_default" {
			continue
		}
		if variant.ID == excludedVariantID {
			continue
		}

		adjustment := int64(0)
		if variant.PriceAdjusment != nil {
			adjustment = *variant.PriceAdjusment
		}

		spaceID := *variant.VariantSpaceID
		minimum, exists := minimumAdjustmentBySpace[spaceID]
		if !exists || adjustment < minimum {
			minimumAdjustmentBySpace[spaceID] = adjustment
		}
	}

	if changedVariant != nil && changedVariant.VariantSpaceID != nil {
		adjustment := int64(0)
		if changedVariant.PriceAdjusment != nil {
			adjustment = *changedVariant.PriceAdjusment
		}

		spaceID := *changedVariant.VariantSpaceID
		minimum, exists := minimumAdjustmentBySpace[spaceID]
		if !exists || adjustment < minimum {
			minimumAdjustmentBySpace[spaceID] = adjustment
		}
	}

	adjustments := make([]int64, 0, len(minimumAdjustmentBySpace))
	for _, adjustment := range minimumAdjustmentBySpace {
		adjustments = append(adjustments, adjustment)
	}

	if priceError := checkVariantAdjustmentPrice(*defaultVariant.PriceAdjusment, adjustments); priceError != "" {
		return priceError, nil
	}

	return "", nil
}
