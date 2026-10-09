package service

import (
	"strconv"
	"strings"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	repo         repository.UserRepository
	perms        *helper.PermissionSet
	appValidator *helper.AppValidator
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet, appValidator *helper.AppValidator) *UserService {
	return &UserService{repo: repo, perms: perms, appValidator: appValidator}
}

func (h *UserService) ListAll(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	// format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	// if err != nil {
	// 	return err
	// }

	q := helper.ParseListQuery(c)
	// q := helper.ParseCursorQuery(c)

	users, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.OkList(c, "berhasil mendapatkan semua user", users, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: CountTotalPages(total, q.Limit),
		Total:      total,
	})

	// rows, err := h.repo.FindAfterCursor(ctx, q)
	// if err != nil {
	// 	return helper.Internal(err)
	// }

	// Baris tambahan hasil limit+1 dipotong di sini. Ia hanya penanda bahwa
	// masih ada halaman berikutnya, bukan bagian dari halaman ini.
	// hasMore := len(rows) > q.Limit
	// if hasMore {
	// 	rows = rows[:q.Limit]
	// }

	// if format == helper.FormatCSV {
	// 	return helper.WriteUsersCSV(c, rows)
	// }

	// meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	// if hasMore && len(rows) > 0 {
	// 	last := rows[len(rows)-1]
	// 	meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	// }

	// return helper.SuccessCursor(c, "daftar user berhasil diambil", rows, meta)
}

func (h *UserService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	current, ok := helper.CurentUser(c)
	if !ok {
		return helper.Unauthorized("can't verify your identity")
	}

	if !CanAccess(current, id, h.perms, "user:read:any") {
		return helper.Forbidden("anda tidak memiliki hak untuk mengakses pengguna ini")
	}

	user, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateErr(err, "user")
	}

	return helper.Ok(c, "berhasil mendapatkan user", user)
}

func (h *UserService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	var req model.CreateUserReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("json is not valid")
	}

	req.Username = strings.TrimSpace(req.Username)

	// if errs := ValidateCreateUser(req); len(errs) > 0 {
	// 	return helper.Validation(errs)
	// }

	if errs := helper.ValidateStruct(req, *h.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	new, err := h.repo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
	})
	if err != nil {
		return translateErr(err, "user")
	}

	return helper.Created(c, "user successfully created!", new, "/api/v1/users/"+strconv.Itoa(new.ID))
}

func (h *UserService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.ReplaceUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	current, ok := helper.CurentUser(c)
	if !ok {
		return helper.Unauthorized("can't verify your identity")
	}

	if !CanAccess(current, id, h.perms, "user:update:any") {
		return helper.Forbidden("anda tidak dapat hak untuk menggantikan data pengguna ini")
	}

	user, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return helper.NotFound("user tidak ditemukan")
	}

	// if errs := ValidateReplaceUser(req); len(errs) > 0 {
	// 	return helper.Validation(errs)
	// }

	if errs := helper.ValidateStruct(req, *h.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	updated, err := h.repo.Update(ctx, user)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, "berhasil memperbarui data pengguna secara penuh", updated)
}

func (h *UserService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	current, ok := helper.CurentUser(c)
	if !ok {
		return helper.Unauthorized("can't verify your identity")
	}

	if !CanAccess(current, id, h.perms, "user:update:any") {
		return helper.Forbidden("anda tidak dapat hak untuk memperbarui data pengguna ini")
	}

	user, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateErr(err, "user")
	}

	if IsEmptyPatchUser(req) {
		return helper.BadRequest("Request Body Kosong!")
	}

	if errs := helper.ValidateStruct(req, *h.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}
	updated := ApplyPatchUser(user, req)

	updated_user, err := h.repo.Update(ctx, updated)
	if err != nil {
		return translateErr(err, "user")
	}

	return helper.Ok(c, "berhasil memperbarui user", updated_user)
}

func (h *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("can't delete user: ID Invalid")
	}

	current, ok := helper.CurentUser(c)
	if !ok {
		return helper.Unauthorized("can't verify your identity")
	}

	if current.UserID == id {
		return helper.Forbidden("tidak memiliki hak untuk menghapus user milik diri sendiri")
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		return translateErr(err, "user")
	}

	return helper.NoContent(c)
}

func (h *UserService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("can't find user: ID Invalid")
	}

	var userAssignRole model.AssignRoleRequest
	if err := c.BodyParser(&userAssignRole); err != nil {
		return helper.BadRequest("JSON invalid")
	}

	errs := helper.ValidateStruct(userAssignRole, *h.appValidator)
	if len(errs) > 0 {
		return helper.Validation(errs)
	}

	current, ok := helper.CurentUser(c)
	if !ok {
		return helper.Unauthorized("can't verify your identity")
	}

	errs = ValidateAssignRole(current, id, userAssignRole, h.perms)
	if len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := h.repo.UpdateRole(ctx, id, userAssignRole.Role)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, "berhasil mengubah role", result)
}
