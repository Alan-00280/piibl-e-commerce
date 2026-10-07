package model

type Stores struct {
	ID          int    `json:"id"`
	OwnerID     int    `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	IsActive    bool   `json:"is_active"`
}

type CreateStoresReq struct {
	OwnerID     int    `json:"owner_id" validate:"required,number,min=1"`
	Name        string `json:"name" validate:"required,min=3,max=45,alphanumspace"`
	Description string `json:"description" validate:"required,min=3,max=300"`
}

type PatchStoresReq struct {
	Name        *string `json:"name,omitempty" validate:"omitnil,min=3,max=45,alphanumspace"`
	Description *string `json:"description,omitempty" validate:"omitnil,min=3,max=300"`
	IsActive    *bool   `json:"is_active,omitempty" validate:"omitnil,boolean"`
}

type StoreFilter struct {
	OwnerID int `json:"owner_id"`
}
