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
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page"),
		Limit:  c.QueryInt("limit"),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
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

	if raw := c.Query("is_active"); raw != "" {
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

	return q
}
