package model

// ROLE
//
// Tipe Role mendefinisikan role apa saja yang mungkin di sistem
// Konstanta Role hanya terdiri dari:
// RoleAdmin, RoleCustomer, RoleTenant, RoleGuest
type Role string

const (
	RoleAdmin    Role = "ADMINISTRATOR"
	RoleCustomer Role = "CUSTOMER"
	RoleTenant   Role = "TENANT"
	RoleGuest    Role = "GUEST"
)

// PRODUCT STATUS
//
// Tipe ProdStat mendefinisikan status suatu produk
// apa saja yang mungkin di sistem
// Konstanta ProdStat hanya terdiri dari:
// ProductStatActive, ProductStatInactive
type ProductStat string

const (
	ProductStatActive   ProductStat = "ACTIVE"
	ProductStatInactive ProductStat = "INACTIVE"
)

// ORDER STATUS
//
// Tipe OrderStat mendefinisikan status suatu order
// apa saja yang mungkin di sistem
// Konstanta OrderStat hanya terdiri dari:
// OrderStatusCreated, OrderStatusCompleted, OrderStatusCancelled
type OrderStat string

const (
	OrderStatusCreated   OrderStat = "CREATED"
	OrderStatusCompleted OrderStat = "COMPLETED"
	OrderStatusCancelled OrderStat = "CANCELLED"
)
