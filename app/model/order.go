package model

import "time"

// untuk per satu customer (immutable)
// (satu customer bisa beberapa order dari berbagai store)
type OrderGroup struct {
	ID         int         `json:"id"`
	CustomerID int         `json:"customer_id"`
	CreatedAt  time.Time   `json:"created_at"`
	Orders     []OrderItem `json:"order_items"`
}

// order untuk dibaca per satu store
// semua immutable kecuali status
type Order struct {
	ID           int         `json:"id"`
	OrderGroupID int         `json:"order_group_id"`
	CustomerID   int         `json:"customer_id"`
	StoreID      int         `json:"store_id"`
	Status       OrderStat   `json:"status"` // non-immutable
	Total        int64       `json:"total"`
	CreatedAt    time.Time   `json:"created_at"`
	OrderItems   []OrderItem `json:"order_items"`
}

// Immutable Order Item
type OrderItem struct {
	ID              int    `json:"id"`
	OrderID         int    `json:"order_id"`
	ProductID       int    `json:"product_id"`
	ProductName     string `json:"product_name"`
	VariantText     string `json:"variant_text"`
	PriceAtPurchase int64  `json:"price_at_purchase"` // Harga saat dibeli
	Quantity        int    `json:"quantity"`
	Subtotal        int64  `json:"subtotal"`
}

// Each Purchased Products
// With it's variant text
type ProductToBeGrouped struct {
	Product
	VariantText     string
	PriceAtPurchase int64
	Quantity        int
}

// CHECKOUT REQ STRUCT
//
// Satu request langsung membuat
// OrderGroup sampai OrderItem
type CheckoutReq struct {
	Items []CheckoutItem `json:"items" validate:"required"`
}

type CheckoutItem struct {
	ProductID  int   `json:"product_id" validate:"required,min=1"`
	VariantIDs []int `json:"variant_ids" validate:"required,min=1,max=500,dive,min=1"` //variant_ids BELUM dibatasi jumlahnya dengan tiap elemen minimal 1
	Quantity   int   `json:"quantity" validate:"required,number,min=1,max=100000"`
}

// REVIEWS
type Review struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	CustomerID  int       `json:"customer_id"`
	VariantText string    `json:"variant_text"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReviewCreateReq struct {
	OrderItemID int     `json:"id" validate:"required,number,min=1"`
	Rating      int     `json:"rating" validate:"required,number,min=1,max=5"`
	Comment     *string `json:"comment,omitempty" validate:"omitnil,min=3,max=300"`
}
