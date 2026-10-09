package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
)

type OrderService struct {
	repo         repository.OrderRepository
	appValidator *helper.AppValidator
}

func NewOrderService(repo repository.OrderRepository, appValidator *helper.AppValidator) *OrderService {
	return &OrderService{
		repo:         repo,
		appValidator: appValidator,
	}
}

// ========================
// CHECKOUT
// ========================
// POST /orders/checkout
//
//		roles: customer
//		desc: Service untuk menangani endpoint pemesanan (Checkout)
//		main process:
//			1. Validasi JSON Metadata
//			2. Validasi JSON setiap Item
//			3. Koleksi semua productIDs dan variantIDs diurutkan
//			4. Ambil semua products dari database menurut productIDs
//			   4.1. Pengecekan apakah ditemukan produk
//			   4.2. Pengecekan apakah produk aktif
//			   4.3. Pegecekan harga _default
//			5. Ambil semua variants dari database menurut variantIDs
//			   5.1. Pengecekan apakah variant ditemukan di database
//			6. Iterasi setiap Items checkout (Setiap Produk):
//			   6.1. Ambil product dan variannya dari products dan variants yang diambil dari database berdasarkan item per checkout
//			   6.2. Pengecekan variants terpilih berdasarkan spaces nya (using: checkChoosenVariants)
//			   6.3. Hitung harga satu item (produk dibeli) berdasarkan varian yang dipilih
//			   6.4. Pemeriksaan awal apakah stok cukup dan aktif
//		       6.5. Ambil spaces dari produk berdasarkan spaces yang dipilih di Item
//		       6.7. Masukkan ke groupedItems
//		    7. Buat order yaitu pesanan dengan grouping Item per Toko
//		    8. Kalkulasi Total Subtotal untuk satu ORDER (Satu TOKO)
//		    9. Lakukan Checkout ke database ==> {Dua Proses Utama: 1. Pengurangan Stok; 2.
//	        Pencatatan Order;}
//		rules:
//			1. Minimal 1 item dipesan
//			2. Variant harus uniques per produk
//			3. Harga total tidak dapat di bawah 100
func (s *OrderService) Checkout(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	var req model.CheckoutReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}
	if len(req.Items) == 0 {
		return helper.BadRequest("minimal satu item harus dipesan")
	}
	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	for index, item := range req.Items {
		if errs := helper.ValidateStruct(item, *s.appValidator); len(errs) > 0 {
			fields := make(map[string]string, len(errs))
			for field, message := range errs {
				fields[fmt.Sprintf("items[%d].%s", index, field)] = message
			}
			return helper.Validation(fields)
		}

		seenVariantIDs := make(map[int]struct{}, len(item.VariantIDs))
		for _, variantID := range item.VariantIDs {
			if _, exists := seenVariantIDs[variantID]; exists {
				return helper.BadRequest(fmt.Sprintf("items[%d] memiliki variant_id duplikat", index))
			}
			seenVariantIDs[variantID] = struct{}{}
		}
	}

	productIDs, variantIDs := checkoutRequestIDs(req.Items)
	products, err := s.repo.FindCheckoutProducts(ctx, productIDs)
	if err != nil {
		return helper.Internal(err)
	}
	if len(products) != len(productIDs) {
		return helper.NotFound("satu atau lebih produk tidak ditemukan")
	}
	for _, productID := range productIDs {
		product := products[productID]
		if product.Product.Status == model.ProductStatInactive {
			return helper.BadRequest(fmt.Sprintf("produk %d tidak aktif", productID))
		}
		if !product.HasBasePrice {
			return helper.Internal(fmt.Errorf("product %d has no _default variant", productID))
		}
	}

	spacesByProduct, err := s.repo.FindCheckoutVariantSpaces(ctx, productIDs)
	if err != nil {
		return helper.Internal(err)
	}
	variantsByID, err := s.repo.FindCheckoutVariants(ctx, variantIDs)
	if err != nil {
		return helper.Internal(err)
	}
	if len(variantsByID) != len(variantIDs) {
		return helper.BadRequest("satu atau lebih varian tidak valid atau tidak ditemukan untuk produk tersebut")
	}

	groupedItems := make([]*model.ProductToBeGrouped, 0, len(req.Items))
	for index, checkoutItem := range req.Items {
		product := products[checkoutItem.ProductID]
		choosens := make([]model.ProductVariant, 0, len(checkoutItem.VariantIDs))
		for _, variantID := range checkoutItem.VariantIDs {
			variant := variantsByID[variantID]
			if variant.ProductID != checkoutItem.ProductID {
				return helper.BadRequest(fmt.Sprintf(
					"items[%d]: satu atau lebih varian tidak valid atau tidak ditemukan untuk produk ini",
					index,
				))
			}
			choosens = append(choosens, variant)
		}

		variantSpaces := spacesByProduct[checkoutItem.ProductID]
		if message := checkChoosenVariants(variantSpaces, choosens); message != "" {
			return helper.BadRequest(fmt.Sprintf("items[%d]: %s", index, message))
		}

		unitPrice := calcOrderItemSubtotal(product.BasePrice, choosens)
		if unitPrice < 100 {
			return helper.BadRequest(fmt.Sprintf("items[%d]: harga produk setelah penyesuaian varian tidak boleh di bawah 100", index))
		}
		if !canBePurchased(product.Product, checkoutItem.Quantity) {
			return helper.Conflict(fmt.Sprintf("stok produk %d tidak mencukupi atau produk tidak aktif", checkoutItem.ProductID))
		}

		selectedSpaces := make([]model.ProductVariantSpace, 0, len(choosens))

		selectedSpaceIDs := make(map[int]struct{}, len(choosens))
		for _, variant := range choosens {
			if variant.VariantSpaceID != nil {
				selectedSpaceIDs[*variant.VariantSpaceID] = struct{}{}
			}
		}

		for _, space := range variantSpaces.VariantSpaces {
			if _, selected := selectedSpaceIDs[space.ID]; selected {
				selectedSpaces = append(selectedSpaces, space)
			}
		}

		groupedItems = append(groupedItems, &model.ProductToBeGrouped{
			Product:         product.Product,
			VariantText:     createChoosenVariantText(choosens, selectedSpaces),
			PriceAtPurchase: unitPrice,
			Quantity:        checkoutItem.Quantity,
		})
	}

	orders := groupItemPerStore(groupedItems, currentUser.UserID)
	for orderIndex := range orders {
		orderItemPointers := make([]*model.OrderItem, 0, len(orders[orderIndex].OrderItems))
		for itemIndex := range orders[orderIndex].OrderItems {
			orderItemPointers = append(orderItemPointers, &orders[orderIndex].OrderItems[itemIndex])
		}
		_, calculatedOrder := countTotalSubtotal(orderItemPointers, &orders[orderIndex])
		orders[orderIndex] = *calculatedOrder
	}

	orderGroup, err := s.repo.CreateCheckout(ctx, currentUser.UserID, req.Items, orders)
	if err != nil {
		if errors.Is(err, repository.ErrStockConflict) {
			return helper.Conflict("stok produk tidak mencukupi atau produk sudah tidak aktif")
		}
		return translateErr(err, "checkout")
	}

	return helper.Created(c, "Pesanan berhasil dibuat", orderGroup, "/api/v1/orders/checkout")
}

func checkoutRequestIDs(items []model.CheckoutItem) ([]int, []int) {
	productSet := make(map[int]struct{})
	variantSet := make(map[int]struct{})
	for _, item := range items {
		productSet[item.ProductID] = struct{}{}
		for _, id := range item.VariantIDs {
			variantSet[id] = struct{}{}
		}
	}

	productIDs := make([]int, 0, len(productSet))
	for id := range productSet {
		productIDs = append(productIDs, id)
	}
	sort.Ints(productIDs)

	variantIDs := make([]int, 0, len(variantSet))
	for id := range variantSet {
		variantIDs = append(variantIDs, id)
	}
	sort.Ints(variantIDs)

	return productIDs, variantIDs
}

// =================
// ORDERS
// =================

// GET /orders
// roles: all except guest (or in other words, any authenticated user)
// Customers see only their orders; tenants see orders for their stores.
func (s *OrderService) ListAll(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	query := helper.ParseListQuery(c)
	orders, total, err := s.repo.FindAll(ctx, query, currentUser)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.OkList(c, "berhasil mendapatkan daftar order", orders, &model.Meta{
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: CountTotalPages(total, query.Limit),
		Total:      total,
	})
}

// GET /orders/:id
// roles: authenticated user
// Customers can read their own orders; tenants can read orders from their stores.
func (s *OrderService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	orderID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	order, err := s.repo.FindByID(ctx, orderID, currentUser)
	if err != nil {
		return translateErr(err, "order")
	}

	return helper.Ok(c, "berhasil mendapatkan detail order", order)
}

// PATCH /orders/:id/status
// role: tenant
// Only the owning tenant can transition an order from CREATED to COMPLETED or CANCELLED.
func (s *OrderService) UpdateStatus(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqContext(c)
	defer cancel()

	orderID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id invalid")
	}

	currentUser, err := checkGetCurrentUser(c)
	if err != nil {
		return err
	}

	var req model.OrderStatusUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("JSON invalid")
	}
	if errs := helper.ValidateStruct(req, *s.appValidator); len(errs) > 0 {
		return helper.Validation(errs)
	}

	order, err := s.repo.FindByID(ctx, orderID, currentUser)
	if err != nil {
		return translateErr(err, "order")
	}
	if message := checkChangeOrderStatus(order.Status, req.Status); message != "" {
		return helper.Conflict(message)
	}

	updatedOrder, err := s.repo.UpdateTenantOrderStatus(ctx, orderID, currentUser.UserID, req.Status)
	if err != nil {
		if errors.Is(err, repository.ErrOrderStatusConflict) {
			return helper.Conflict("status order sudah berubah dan tidak dapat diubah lagi")
		}
		return translateErr(err, "order")
	}

	return helper.Ok(c, "status order berhasil diperbarui", updatedOrder)
}
