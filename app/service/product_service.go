package service

import (
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
)

type ProductService struct {
	repo         repository.ProductRepository
	perms        *helper.PermissionSet
	appValidator *helper.AppValidator
}

func NewProductService(
	repo repository.ProductRepository,
	perms *helper.PermissionSet,
	appValidator *helper.AppValidator,
) *ProductService {
	return &ProductService{
		repo:         repo,
		perms:        perms,
		appValidator: appValidator,
	}
}

// GET /products
// roles: any
// perms: -
// rules:
//  1. Jika inactive jangan tampilkan (kecuali pemilik produk)
//  2. Stok kosong tetap tampilkan

// GET /products/:id
// roles: any
// perms: -
// rules:
//  1. Jika inactive jangan tampilkan (kecuali pemilik produk)
//  2. Stok kosong tetap tampilkan

// POST /products
// roles: tenant
// perms:
// rules:
//   1. _default dikelola sistem
//   2. _default wajib ada harga

// PATCH /products/:id
// roles: tenant
// perms:
// rules:
//   1. Produk milik sendiri

// DELETE /products/:id (menonaktifkan)
// roles: tenant
// perms:
// rules:
//   1. Produk milik sendiri
//   2. Produk tidak permanently dihapus

// PATCH /products/:id/stock
// roles: tenant
// perms:
// rules:
//   1. Produk milik sendiri
//   2. Stok tidak boleh negatif

// POST /products/:id/variant-spaces
// roles: tenant
// perms:
// rules:
//   1. Produk milik sendiri

// PATCH variant-spaces/:id
// roles: tenant
// perms
// rules:
//   1. Produk sendiri

// DELETE /variant-spaces/:id
// roles: tenant, admin
// perms:
//   tenant -> ownership check
//   admin  -> variant:manage:any

// POST variant-spaces/:id/variants
// roles: tenant, admin
// perms:
//   tenant -> ownership check
//   admin  -> variant:manage:any
// rules:
//   1. Produk harus milik tenant
//   2. Variant yang dibuat adalah variant biasa
//   3. Tidak boleh membuat variant bernama _default

// PATCH variant-spaces/:id/variants/:id
// roles: tenant, admin
// perms:
//   tenant -> ownership check
//   admin  -> variant:manage:any
// rules:
//   1. Produk harus milik tenant
//   2. Variant harus berada di product/space tersebut
//   3. _default tidak boleh diperlakukan sebagai variant biasa

// DELETE variant-spaces/:id/variants/:id
// roles: tenant, admin
// perms:
//   tenant -> ownership check
//   admin  -> variant:manage:any
// rules:
//   1. Produk harus milik tenant
//   2. _default tidak boleh dihapus
//   3. Variant biasa boleh dihapus permanen
