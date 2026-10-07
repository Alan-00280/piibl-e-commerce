package model

import "time"

// PRODUCTS
type Product struct {
	ID          int         `json:"id"`
	StoreID     int         `json:"store_id"`
	CategoryID  int         `json:"category_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Stock       int         `json:"stock"`
	Status      ProductStat `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ProductCreateReq struct {
	StoreID     int         `json:"store_id" validate:"required,number,min=1"`
	CategoryID  int         `json:"category_id" validate:"required,number,min=1"`
	Name        string      `json:"name" validate:"required,min=3,max=100,alphanumspace"`
	Description string      `json:"description" validate:"required,min=3,max=300"`
	Stock       int         `json:"stock" validate:"number,min=0,max=100000"`
	Status      ProductStat `json:"status" validate:"required"`
}

type ProductPatchReq struct {
	CategoryID  *int         `json:"category_id,omitempty" validate:"omitnil,number,min=1"`
	Name        *string      `json:"name,omitempty" validate:"omitnil,min=3,max=100,alphanumspace"`
	Description *string      `json:"description,omitempty" validate:"omitnil,min=3,max=300"`
	Stock       *int         `json:"stock,omitempty" validate:"omitnil,number,min=0,max=100000"`
	Status      *ProductStat `json:"status,omitempty" validate:"omitnil"`
}

type ProductDeactivateReq struct {
	Status ProductStat `json:"status" validate:"required"`
}

type ProductUpdateStockReq struct {
	Stock int `json:"stock" validate:"number,min=0,max=100000"`
}

type ProductFilter struct {
	StoreID    int
	CategoryID int
}

type ProductVariantSpaces struct {
	ProductID     int
	VariantSpaces []ProductVariantSpace
}

// PRODUCT VARIANTS
type ProductVariantSpace struct {
	ID              int              `json:"id"`
	ProductID       int              `json:"product_id"`
	Name            string           `json:"name"`
	IsMandatory     bool             `json:"is_mandatory"`
	ProductVariants []ProductVariant `json:"variants"`
}

type ProductVariantSpaceCreateReq struct {
	ProductID   int    `json:"product_id" validate:"required,number,min=1"`
	Name        string `json:"name" validate:"required,min=3,max=20,alphanumspace"`
	IsMandatory *bool  `json:"is_mandatory" validate:"required,boolean"`
}

type ProductVariantSpacePatchReq struct {
	Name        *string `json:"name" validate:"omitnil,min=3,max=20,alphanumspace"`
	IsMandatory *bool   `json:"is_mandatory" validate:"omitnil,boolean"`
}

type ProductVariant struct {
	ID             int    `json:"id"`
	ProductID      int    `json:"product_id"`
	VariantSpaceID *int   `json:"space_id,omitempty"`
	Name           string `json:"name"`
	PriceAdjusment *int64 `json:"price,omitempty"`
}

type ProductVariantItem struct {
	Name           string `json:"name" validate:"required,min=3,max=50"`
	PriceAdjusment *int64 `json:"price_adjustment,omitempty" validate:"omitnil,number,min=-1000000000,max=100000000000"`
}

type ProductVariantCreateReq struct {
	ProductID           int                `json:"product_id" validate:"required,number,min=1"`
	VariantSpaceID      *int               `json:"space_id,omitempty" validate:"omitnil,number,min=1"`
	ProductVariantItems ProductVariantItem `json:"product_variant_items"`
}

type ProductVariantPatchReq struct {
	Name           *string `json:"name,omitempty" validate:"omitnil,min=3,max=50"`
	PriceAdjusment *int64  `json:"price,omitempty" validate:"omitnil,number,min=-100000000000,max=100000000000"`
}
