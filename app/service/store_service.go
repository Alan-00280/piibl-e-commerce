package service

import (
	"strconv"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
)

type StoreService struct {
	repo         repository.StoreRepository
	perms        *helper.PermissionSet
	appValidator *helper.AppValidator
}

func NewStoreService(
	repo repository.StoreRepository,
	perms *helper.PermissionSet,
	appValidator *helper.AppValidator,
) *StoreService {
	return &StoreService{
		repo:         repo,
		perms:        perms,
		appValidator: appValidator,
	}
}

// GET /stores
//
//	role: any
//	perms: -
//	rules:
//	  1. Jika deactivated (jangan tampilkan) [implementasi di repository]
func (s *StoreService) ListAll(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)
	stores, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.OkList(c, "berhasil mendapatkan semua store", stores, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: CountTotalPages(total, q.Limit),
		Total:      total,
	})
}

// GET /stores/:id
//
//	role: any
//	perms: -
//	rules:
//	  1. Jika deactivated jangan tampilkan
func (s *StoreService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	if store.IsActive == false {
		currentUser, ok := helper.CurentUser(c)
		if !ok {
			return helper.Forbidden("store sedang inactive")
		}

		if currentUser.UserID != store.TenantID {
			return helper.Forbidden("anda tidak memiliki hak atas inactive store ini")
		}
	}

	return helper.Ok(c, "store ditemukan", store)
}

// GET /my/stores
//
//	role: tenant
//	desc: mendapatkan store milik tenant yang sedang terautentikasi
func (s *StoreService) GetStoreTenant(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	currentUser, _ := checkGetCurrentUser(c)

	store, err := s.repo.FindByTenantID(ctx, currentUser.UserID)
	if err != nil {
		return translateErr(err, "store")
	}

	return helper.Ok(c, "store ditemukan", store)
}

// POST /stores
//
//	role: tenant
//	perms: store:create
//	rules:
//	  1. Satu tenant maksimal satu store (baik aktif maupun tidak, maksimal satu.)
func (s *StoreService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	currentUser, _ := checkGetCurrentUser(c)

	var req model.CreateStoresReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("json is not valid")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	if _, err := s.repo.FindByTenantID(ctx, currentUser.UserID); err == nil {
		return helper.Conflict("tenant sudah memiliki store")
	}

	new, err := s.repo.Create(ctx, model.Stores{
		TenantID:    currentUser.UserID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	})
	if err != nil {
		return translateErr(err, "store")
	}

	return helper.Created(c, "store berhasil dibuat", new, "/api/v1/stores/"+strconv.Itoa(new.ID))
}

// PATCH /stores/:id
//
//	role: tenant
//	perms: store:update:any
//	rules:
//	  1. Hanya pemilik toko
//	  2. pemilik Perms
func (s *StoreService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.PatchStoresReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	currentUser, _ := checkGetCurrentUser(c)

	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErr(err, "store")
	}

	if !CanAccess(currentUser, store.TenantID, s.perms, "store:update:any") {
		return helper.Forbidden("Anda tidak memiliki hak atas store ini")
	}

	if IsEmptyPatchStore(req) {
		return helper.BadRequest("body kosong")
	}

	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	updated := ApplyPatchStore(store, req)
	store_updated, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateErr(err, "store")
	}

	return helper.Ok(c, "data store berhasil diperbarui", store_updated)
}

// DELETE /stores/:id (deactivating)
//
//	role: tenant
//	perms:
//	rules:
//	  1. Hanya pemilik toko
//	  2. Pemilik perms
func (s *StoreService) Deactivate(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	currentUser, _ := checkGetCurrentUser(c)

	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateErr(err, "store")
	}

	if !CanAccess(currentUser, store.TenantID, s.perms, "store:delete") {
		return helper.Forbidden("Anda tidak memiliki hak atas store ini")
	}

	if err := s.repo.Deactivate(ctx, id); err != nil {
		translateErr(err, "store")
	}

	return helper.NoContent(c)
}
