package model

type Stores struct {
	ID          int    `json:"id"`
	TenantID    int    `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	IsActive    bool   `json:"is_active"`
}

type CreateStoresReq struct {
	Name        string `json:"name" validate:"required,min=3,max=45,alphanumspace"`
	Description string `json:"description" validate:"required,min=3,max=300"`
}

type PatchStoresReq struct {
	Name        *string `json:"name,omitempty" validate:"omitnil,min=3,max=45,alphanumspace"`
	Description *string `json:"description,omitempty" validate:"omitnil,min=3,max=300"`
	IsActive    *bool   `json:"is_active,omitempty" validate:"omitnil,boolean"`
}

type StoreFilter struct {
	TenantID *int
}
