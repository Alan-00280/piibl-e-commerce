package model

import "time"

// Amplop baku untuk semua respons
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type WebResponseCursor struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
	Meta    *CursorMeta `json:"meta,omitempty"`
	Errors  any         `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type ErrorRespone struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	*UserFilter
	*StoreFilter
	*ProductFilter
	*OrderFilter
}

type CursorQuery struct {
	Search   string
	Limit    int
	IsActive *bool
	After    *Cursor
	*UserFilter
	*StoreFilter
	*ProductFilter
	*OrderFilter
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
