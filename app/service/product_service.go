package service

import (
	"fmt"
	"strconv"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
)

type ProductService struct {
	repo         repository.ProductRepository
	storeRepo    repository.StoreRepository
	perms        *helper.PermissionSet
	appValidator *helper.AppValidator
}

func NewProductService(
	repo repository.ProductRepository,
	storeRepo repository.StoreRepository,
	perms *helper.PermissionSet,
	appValidator *helper.AppValidator,
) *ProductService {
	return &ProductService{
		repo:         repo,
		storeRepo:    storeRepo,
		perms:        perms,
		appValidator: appValidator,
	}
}

var defaultCategory int = 6 // ADA DI DATABASE category id=6

// ==================
// PRODUCT
// ==================

// GET /products
//
//	roles: any
//	perms: -
//	rules:
//	  1. Jika inactive jangan tampilkan
//	  2. Stok kosong tetap tampilkan
func (s *ProductService) ListAll(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)
	products, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.OkList(c, "berhasil mendapatkan semua produk", products, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: CountTotalPages(total, q.Limit),
		Total:      total,
	})
}

// GET /products/:id
//
//	roles: any
//	perms: -
//	rules:
//	  1. Jika inactive jangan tampilkan
//	  2. Stok kosong tetap tampilkan
func (s *ProductService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErr(err, "product")
	}

	if product.Product.Status == model.ProductStatInactive {
		return helper.Forbidden("produk sedang inactive")
	}

	variantSpaces, err := s.repo.FindVariantSpacesByProductID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	data := struct {
		Product       model.Product               `json:"Product"`
		BasePrice     int64                       `json:"BasePrice"`
		VariantSpaces []model.ProductVariantSpace `json:"variant_spaces"`
	}{
		Product:       product.Product,
		BasePrice:     product.BasePrice,
		VariantSpaces: variantSpaces.VariantSpaces,
	}

	return helper.Ok(c, "product ditemukan", data)
}

// GET /my/products
//
//	roles: tenant
//	desc: Untuk mendapatkan semua produk baik aktif maupun tidak milik store
//	rules:
//	   1. tenant sudah memiliki store
func (s *ProductService) ProductStore(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	currentUser, _ := checkGetCurrentUser(c)
	store, _ := getTenantStores(currentUser.UserID, s.storeRepo, ctx)

	products, err := s.repo.FindByStoreID(ctx, store.ID)
	if err != nil {
		return translateErr(err, "store")
	}

	return helper.OkList(c, "berhasil mendapatkan semua produk", products, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: CountTotalPages(len(products), q.Limit),
		Total:      len(products),
	})
}

// POST /products
// roles: tenant
// perms:
// rules:
//  1. _default dikelola sistem
//  2. _default wajib ada harga
func (s *ProductService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	var req model.ProductCreateReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("json is not valid")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	store, err := getTenantStores(currentUser.UserID, s.storeRepo, ctx)
	if err != nil {
		return err
	}

	productCategory := defaultCategory
	if req.CategoryID != nil {
		productCategory = *req.CategoryID
	}

	new, err := s.repo.Create(ctx, model.Product{
		StoreID:     store.ID,
		CategoryID:  productCategory,
		Name:        req.Name,
		Description: req.Description,
		Stock:       req.Stock,
		Status:      req.Status,
	}, req.BasePrice)
	if err != nil {
		return translateErr(err, "product")
	}

	return helper.Created(c, "product berhasil dibuat", new, "/api/v1/products/"+strconv.Itoa(new.Product.ID))
}

// PATCH /products/:id
// roles: tenant
// desc: untuk mengubah data produk tanpa mengubah harga dasar
// perms: product:update:any
// rules:
//  1. Produk milik sendiri
func (s *ProductService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ProductPatchReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, id)
	if err != nil {
		return translateErr(err, "product")
	}

	if !CanAccess(currentUser, productTenantID, s.perms, "product:update:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErr(err, "product")
	}

	if IsEmptyPatchProduct(req) {
		return helper.BadRequest("body kosong")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	updated := ApplyPatchProduct(product.Product, req)
	product_updated, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateErr(err, "store")
	}

	return helper.Ok(c, "data product berhasil diperbarui", product_updated)
}

// DELETE /products/:id (menonaktifkan)
// roles: tenant
// perms:
// rules:
//  1. Produk milik sendiri
//  2. Produk tidak permanently dihapus
func (s *ProductService) Deactivate(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, id)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "product:update:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	if err := s.repo.Deactivate(ctx, id); err != nil {
		return translateErr(err, "product")
	}

	return helper.NoContent(c)
}

// PATCH /products/:id/stock
// roles: tenant
// perms:
// rules:
//  1. Produk milik sendiri
//  2. Stok tidak boleh negatif
func (s *ProductService) Restock(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ProductUpdateStockReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, id)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "product:update:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	if err := s.repo.UpdateStock(ctx, id, req.Stock); err != nil {
		return translateErr(err, "product")
	}

	return helper.Ok(c, "berhasil diperbarui", fmt.Sprintf("Stok Baru: %d", req.Stock))
}

// PATCH /products/:id/price
//
//	roles: tenant
//	desc: untuk mengubah harga dasar produk
func (s *ProductService) Reprice(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ProductBasePriceReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	productTenantID, err := s.repo.FindTenantIDByProductID(ctx, id)
	if err != nil {
		return translateErr(err, "product")
	}
	if !CanAccess(currentUser, productTenantID, s.perms, "product:update:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas produk ini")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	defaultVariant, err := s.repo.FindDefaultVariant(ctx, id)
	if err != nil {
		return translateErr(err, "product default variant")
	}
	defaultPrice := req.Baseprice
	defaultVariant.PriceAdjusment = &defaultPrice

	updatedVariant, err := s.repo.UpdateVariant(ctx, defaultVariant)
	if err != nil {
		return translateErr(err, "product default variant")
	}

	return helper.Ok(c, "harga dasar produk berhasil diperbarui", updatedVariant)
}
