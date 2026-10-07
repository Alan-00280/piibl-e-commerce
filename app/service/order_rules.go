package service

import (
	"fmt"
	"strings"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
)

// Menghitung total untuk satu item
// Berdasarkan variannya
func calcOrderItemSubtotal(defaultPrice int64, productVariantChoosen []model.ProductVariant) int64 {
	if len(productVariantChoosen) == 0 {
		return defaultPrice
	}

	var total int64
	total = 0

	for _, productVariant := range productVariantChoosen {
		if productVariant.PriceAdjusment != nil {
			total += *productVariant.PriceAdjusment
		}
	}

	return defaultPrice + total
}

// Memeriksa validitas varian yang dipilih:
//
// Tolak: variant bukan milik produk,
// dua nilai dalam satu ruang, ruang wajib tidak dipilih,
// id ganda, memilih _default secara eksplisit
func checkChoosenVariants(productVariantSpaces model.ProductVariantSpaces, choosens []model.ProductVariant) string {
	if duplicatedSpaceID, yes := duplicateChoosenVariants(choosens); yes {
		if duplicatedSpaceID == -2 {
			return "_default dilarang masuk ke choosen variant!"
		}

		var spaceName string
		for _, productVariantSpace := range productVariantSpaces.VariantSpaces {
			if productVariantSpace.ID == duplicatedSpaceID {
				spaceName = productVariantSpace.Name
			}
		}

		return fmt.Sprintf("ruang varian %s memiliki nilai ganda", spaceName)
	}

	choosenSpaces := make(map[int]struct{})
	productSpaces := make(map[int]bool)
	for _, productSpace := range productVariantSpaces.VariantSpaces {
		productSpaces[productSpace.ID] = productSpace.IsMandatory
	}

	for _, choosen := range choosens {
		if choosen.VariantSpaceID == nil {
			return "_default dilarang masuk ke choosen variant!"
		}

		if choosen.VariantSpaceID != nil {
			if _, exists := productSpaces[*choosen.VariantSpaceID]; !exists {
				return "variant bukan milik produk"
			}

			choosenSpaces[*choosen.VariantSpaceID] = struct{}{}
		}
	}

	var mandatorySpaces []int
	for _, productVariantSpace := range productVariantSpaces.VariantSpaces {
		if productVariantSpace.IsMandatory {
			mandatorySpaces = append(mandatorySpaces, productVariantSpace.ID)
		}
	}

	switch {
	case len(mandatorySpaces) > 0 && len(choosens) == 0:
		return "belum ada ruang yang dipilih"
	case len(mandatorySpaces) == 0 && len(choosens) == 0:
		return ""
	}

	for mandatorySpaceId, isMandatory := range productSpaces {
		if isMandatory {
			if _, exists := choosenSpaces[mandatorySpaceId]; !exists {
				return "terdapat ruang yang belum dipilih"
			}
		}
	}

	return ""
}

// Memeriksa apakah ada varian yang nilainya ganda
// Pada ruang varian yang sama
func duplicateChoosenVariants(choosens []model.ProductVariant) (int, bool) {
	var choosenIDs []int
	for _, choosen := range choosens {
		if choosen.VariantSpaceID != nil {
			choosenIDs = append(choosenIDs, *choosen.VariantSpaceID)
		} else {
			return -2, true
		}
	}

	visited := make(map[int]bool, 0)
	for i := 0; i < len(choosenIDs); i++ {
		if visited[choosenIDs[i]] == true {
			return *choosens[i].VariantSpaceID, true
		} else {
			visited[choosenIDs[i]] = true
		}
	}
	return -1, false
}

// createChoosenVariantText returns chosen variants in the format:
// SPACE-variant|SPACE-variant
// Variant name spaces are replaced with hyphens.
//
// Example: WARNA:Merah-Tua|MODEL:Shoulder-Bag|BAHAN-LUAR:Kulit-Sintetis
func createChoosenVariantText(choosenVariants []model.ProductVariant, choosenSpace []model.ProductVariantSpace) string {
	var parts []string

	variants := make(map[int]string)
	for _, choosenVariant := range choosenVariants {
		if choosenVariant.VariantSpaceID != nil {
			variants[*choosenVariant.VariantSpaceID] = strings.ReplaceAll(choosenVariant.Name, " ", "-")
		}
	}

	for _, choosenSpace := range choosenSpace {
		spaceName := choosenSpace.Name
		spaceName = strings.ToUpper(spaceName)
		spaceName = strings.ReplaceAll(spaceName, " ", "-")

		parts = append(parts, fmt.Sprintf("%s:%s", spaceName, variants[choosenSpace.ID]))
	}

	return strings.Join(parts, "|")
}

// Memeriksa kuantitas sebelum melakukan
// Atomic update di tabel product
func canBePurchased(product model.Product, qty int) bool {
	currentStock := product.Stock

	if product.Status == model.ProductStatInactive {
		return false
	}

	if qty <= 0 {
		return false
	}

	return qty <= currentStock
}

// Mengembalikan nilai stok yang baru
// Dipakai untuk membantu pengatasan kondisi Race Condition
func decreasedStock(product *model.Product, qty int) (model.Product, error) {
	currentStock := product.Stock

	newStock := currentStock - qty

	if newStock < 0 {
		return model.Product{}, fmt.Errorf("Terjadi kesalahan. Nilai stok baru dibawah 0")
	}

	product.Stock = newStock
	return *product, nil
}

// Menghitung harga subtotal per item serta total
func countTotalSubtotal(orderItems []*model.OrderItem, order *model.Order) ([]model.OrderItem, model.Order) {
	var appliedOrderItems []model.OrderItem

	var total int64

	var subtotal int64
	for _, orderItem := range orderItems {
		price := orderItem.PriceAtPurchase
		quantity := orderItem.Quantity

		subtotal = price * int64(quantity)
		orderItem.Subtotal = subtotal

		appliedOrderItems = append(appliedOrderItems, *orderItem)
		total += subtotal
	}

	order.Total = total
	return appliedOrderItems, *order
}

// Memeriksa Pengubahan Status Order
func checkChangeOrderStatus(older model.OrderStat, newer model.OrderStat) string {
	switch {
	case older == model.OrderStatusCreated && newer == model.OrderStatusCancelled:
		return ""
	case older == model.OrderStatusCreated && newer == model.OrderStatusCompleted:
		return ""
	default:
		return fmt.Sprintf("status %s tidak dapat diubah ke %s", older, newer)
	}
}

// Memeriksa apakah order bisa direview
// Input berupa:
//   1. order item yang akan direview
//   2. catatan order dari database
//   3. daftar catatan item order dari database
func canReviews(orderItem model.OrderItem, orderRecord model.Order, orderItemRecords []model.OrderItem) string {
	if orderRecord.Status != model.OrderStatusCompleted {
		return fmt.Sprintf("pesanan masih berstatus %v", orderRecord.Status)
	}

	var orderExists bool
	orderExists = false

	orderProductIDs := make(map[int]struct{})
	for _, orderItemRecord := range orderItemRecords {
		if orderItemRecord.OrderID != orderRecord.ID {
			return "catatan order item tidak sesuai"
		}

		if orderItemRecord.ID == orderItem.ID {
			orderExists = true
		}

		orderProductIDs[orderItemRecord.ProductID] = struct{}{}
	}
	if !orderExists {
		return "order item tidak ada di catatan order"
	}

	if _, exists := orderProductIDs[orderItem.ProductID]; !exists {
		return "produk tidak tercatat di pesanan"
	}

	return ""
}
