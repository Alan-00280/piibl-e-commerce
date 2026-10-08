package helper

import (
	"strconv"
	"strings"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/gofiber/fiber/v2"
)

var allowedSort = map[string]bool{
	"id":       true,
	"username": true,
	"name":     true,
	"price":    true,
	"rating":   true,
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	trueActive := true

	q := model.ListQuery{
		Page:     c.QueryInt("page"),
		Limit:    c.QueryInt("limit"),
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     c.Query("sort", "id"),
		Order:    strings.ToLower(c.Query("order", "asc")),
		IsActive: &trueActive,
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	if q.Order != "desc" {
		q.Order = "asc"
	}

	if raw := c.Query("is_active"); strings.TrimSpace(raw) != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	userFilter := model.UserFilter{
		Role: "",
	}
	if roleFilter := c.Query("role"); strings.TrimSpace(roleFilter) != "" {
		if role, valid := RoleOf(strings.ToUpper(roleFilter)); valid && role != model.RoleAdmin {
			userFilter.Role = role
		}
	}
	q.UserFilter = &userFilter

	storesFilter := model.StoreFilter{}
	if tenantID := c.Query("tenant_id"); strings.TrimSpace(tenantID) != "" {
		if v, err := strconv.Atoi(tenantID); err != nil && v > 0 {
			storesFilter.TenantID = &v
		}
	}
	q.StoreFilter = &storesFilter

	productFilter := model.ProductFilter{
		Status: model.ProductStatActive,
	}
	if StoreID := c.Query("store_id"); strings.TrimSpace(StoreID) != "" {
		if v, err := strconv.Atoi(StoreID); err != nil && v > 0 {
			productFilter.StoreID = &v
		}
	}
	if CategoryID := c.Query("category"); strings.TrimSpace(CategoryID) != "" {
		if v, err := strconv.Atoi(CategoryID); err != nil && v > 0 {
			productFilter.CategoryID = &v
		}
	}
	if MaxPrice := c.Query("maxprice"); strings.TrimSpace(MaxPrice) != "" {
		if v, err := strconv.ParseInt(MaxPrice, 10, 64); err != nil && v > 0 {
			productFilter.PriceMax = &v
		}
	}
	if MinPrice := c.Query("minprice"); strings.TrimSpace(MinPrice) != "" {
		if v, err := strconv.ParseInt(MinPrice, 10, 64); err != nil && v > 0 {
			productFilter.PriceMin = &v
		}

		var zeroInt64 int64
		zeroInt64 = int64(0)
		productFilter.PriceMin = &zeroInt64
	}
	if RatingMax := c.Query("maxrate"); strings.TrimSpace(RatingMax) != "" {
		if v, err := strconv.ParseFloat(RatingMax, 64); err != nil && v > 0 {
			productFilter.MaxRating = &v
		}
	}
	if RatingMin := c.Query("minrate"); strings.TrimSpace(RatingMin) != "" {
		if v, err := strconv.ParseFloat(RatingMin, 64); err != nil && v > 0 {
			productFilter.MinRating = &v
		}
	}
	ProductStatus := c.Query("product_stat")
	if ProductStatus = strings.ToUpper(strings.TrimSpace(ProductStatus)); ProductStatus != "" {
		switch {
		case ProductStatus == string(model.ProductStatActive):
			productFilter.Status = model.ProductStatActive
		case ProductStatus == string(model.ProductStatInactive):
			productFilter.Status = model.ProductStatInactive
		default:
			productFilter.Status = model.ProductStatActive
		}
	}
	q.ProductFilter = &productFilter

	return q
}
