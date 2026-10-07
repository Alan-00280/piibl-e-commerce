package service

import (
	"testing"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
)

func TestProductVariantPricingFlow(t *testing.T) {
	const defaultPrice int64 = 50000

	colorSpaceID := 10
	sizeSpaceID := 20
	foreignSpaceID := 30

	redAdjustment := int64(5000)
	blueAdjustment := int64(7000)
	largeAdjustment := int64(10000)

	red := model.ProductVariant{
		ID: 101, ProductID: 1, VariantSpaceID: &colorSpaceID, Name: "Merah", PriceAdjusment: &redAdjustment,
	}
	blue := model.ProductVariant{
		ID: 102, ProductID: 1, VariantSpaceID: &colorSpaceID, Name: "Biru", PriceAdjusment: &blueAdjustment,
	}
	medium := model.ProductVariant{
		ID: 201, ProductID: 1, VariantSpaceID: &sizeSpaceID, Name: "M",
	}
	large := model.ProductVariant{
		ID: 202, ProductID: 1, VariantSpaceID: &sizeSpaceID, Name: "L", PriceAdjusment: &largeAdjustment,
	}
	explicitDefault := model.ProductVariant{ID: 1, ProductID: 1, Name: "_default"}
	foreignVariant := model.ProductVariant{
		ID: 301, ProductID: 2, VariantSpaceID: &foreignSpaceID, Name: "Hijau",
	}

	spaces := model.ProductVariantSpaces{
		ProductID: 1,
		VariantSpaces: []model.ProductVariantSpace{
			{ID: colorSpaceID, ProductID: 1, Name: "Warna"},
			{ID: sizeSpaceID, ProductID: 1, Name: "Ukuran", IsMandatory: true},
		},
	}
	product := model.Product{
		ID: 1, Name: "Baju", Status: model.ProductStatActive, Stock: 10,
	}

	t.Run("step 1: default price applies but purchase is invalid without mandatory Size", func(t *testing.T) {
		if got := calcOrderItemSubtotal(defaultPrice, nil); got != defaultPrice {
			t.Errorf("default subtotal = %d, want %d", got, defaultPrice)
		}

		if got, want := checkChoosenVariants(spaces, nil), "belum ada ruang yang dipilih"; got != want {
			t.Errorf("empty selection validation = %q, want %q", got, want)
		}
	})

	t.Run("step 2: Red and M add only the Red adjustment to default price", func(t *testing.T) {
		choices := []model.ProductVariant{red, medium}
		if got := checkChoosenVariants(spaces, choices); got != "" {
			t.Fatalf("Red + M validation = %q, want valid selection", got)
		}
		if got, want := calcOrderItemSubtotal(defaultPrice, choices), int64(55000); got != want {
			t.Errorf("Red + M price = %d, want %d", got, want)
		}
	})

	t.Run("step 3: replacing M with L adds both adjustments to default price", func(t *testing.T) {
		choices := []model.ProductVariant{red, large}
		if got := checkChoosenVariants(spaces, choices); got != "" {
			t.Fatalf("Red + L validation = %q, want valid selection", got)
		}
		if got, want := calcOrderItemSubtotal(defaultPrice, choices), int64(65000); got != want {
			t.Errorf("Red + L price = %d, want %d", got, want)
		}
	})

	t.Run("step 4: selecting Red and Blue from the same space is rejected", func(t *testing.T) {
		if got, want := checkChoosenVariants(spaces, []model.ProductVariant{red, blue}), "ruang varian Warna memiliki nilai ganda"; got != want {
			t.Errorf("duplicate color validation = %q, want %q", got, want)
		}
	})

	t.Run("step 5: selecting only Red is rejected because Size is mandatory", func(t *testing.T) {
		if got, want := checkChoosenVariants(spaces, []model.ProductVariant{red}), "terdapat ruang yang belum dipilih"; got != want {
			t.Errorf("missing mandatory selection = %q, want %q", got, want)
		}
	})

	t.Run("additional rule: the default variant cannot be selected explicitly", func(t *testing.T) {
		if got := checkChoosenVariants(spaces, []model.ProductVariant{explicitDefault}); got == "" {
			t.Error("explicit _default selection was accepted, want it rejected")
		}
	})

	t.Run("additional rule: variant from another product is rejected", func(t *testing.T) {
		if got := checkChoosenVariants(spaces, []model.ProductVariant{foreignVariant}); got != "variant bukan milik produk" {
			t.Errorf("foreign variant validation = %q, want %q", got, "variant bukan milik produk")
		}
	})

	t.Run("additional rule: duplicate variant is rejected", func(t *testing.T) {
		if got := checkChoosenVariants(spaces, []model.ProductVariant{red, red}); got != "ruang varian Warna memiliki nilai ganda" {
			t.Errorf("duplicate variant validation = %q, want duplicate-space error", got)
		}
	})

	t.Run("product fixture is active", func(t *testing.T) {
		if product.Status != model.ProductStatActive {
			t.Errorf("product status = %q, want %q", product.Status, model.ProductStatActive)
		}
	})
}

func TestOrderStatusTransitionFlow(t *testing.T) {
	t.Run("step 6: CREATED can transition to COMPLETED", func(t *testing.T) {
		order := model.Order{Status: model.OrderStatusCreated}
		if got := checkChangeOrderStatus(order.Status, model.OrderStatusCompleted); got != "" {
			t.Fatalf("CREATED -> COMPLETED validation = %q, want valid transition", got)
		}
		order.Status = model.OrderStatusCompleted
		if order.Status != model.OrderStatusCompleted {
			t.Errorf("order status = %q, want %q", order.Status, model.OrderStatusCompleted)
		}
	})

	t.Run("step 7: CREATED can transition to CANCELLED", func(t *testing.T) {
		order := model.Order{Status: model.OrderStatusCreated}
		if got := checkChangeOrderStatus(order.Status, model.OrderStatusCancelled); got != "" {
			t.Fatalf("CREATED -> CANCELLED validation = %q, want valid transition", got)
		}
		order.Status = model.OrderStatusCancelled
		if order.Status != model.OrderStatusCancelled {
			t.Errorf("order status = %q, want %q", order.Status, model.OrderStatusCancelled)
		}
	})

	t.Run("step 8: COMPLETED cannot transition to CANCELLED", func(t *testing.T) {
		order := model.Order{Status: model.OrderStatusCompleted}
		if got := checkChangeOrderStatus(order.Status, model.OrderStatusCancelled); got == "" {
			t.Fatal("COMPLETED -> CANCELLED was accepted, want rejection")
		}
		if order.Status != model.OrderStatusCompleted {
			t.Errorf("rejected transition changed order status to %q", order.Status)
		}
	})

	t.Run("step 9: CANCELLED cannot transition back to CREATED", func(t *testing.T) {
		order := model.Order{Status: model.OrderStatusCancelled}
		if got := checkChangeOrderStatus(order.Status, model.OrderStatusCreated); got == "" {
			t.Fatal("CANCELLED -> CREATED was accepted, want rejection")
		}
		if order.Status != model.OrderStatusCancelled {
			t.Errorf("rejected transition changed order status to %q", order.Status)
		}
	})
}

func TestProductPurchaseFlow(t *testing.T) {
	t.Run("step 10: insufficient stock rejects purchase without changing stock", func(t *testing.T) {
		product := model.Product{ID: 1, Status: model.ProductStatActive, Stock: 3}
		const quantity = 4

		if canBePurchased(product, quantity) {
			t.Fatal("purchase of quantity 4 with stock 3 was accepted")
		}
		if product.Stock != 3 {
			t.Errorf("stock after rejected purchase = %d, want 3", product.Stock)
		}

		if _, err := decreasedStock(&product, quantity); err == nil {
			t.Fatal("decreasedStock accepted a purchase that would make stock negative")
		}
		if product.Stock != 3 {
			t.Errorf("stock after failed stock decrement = %d, want 3", product.Stock)
		}
	})

	t.Run("additional rule: non-positive quantities are rejected", func(t *testing.T) {
		for _, quantity := range []int{0, -1} {
			product := model.Product{Status: model.ProductStatActive, Stock: 3}
			if canBePurchased(product, quantity) {
				t.Errorf("purchase quantity %d was accepted, want rejection", quantity)
			}
			if product.Stock != 3 {
				t.Errorf("stock after rejecting quantity %d = %d, want 3", quantity, product.Stock)
			}
		}
	})

	t.Run("additional rule: inactive product cannot be purchased", func(t *testing.T) {
		product := model.Product{Status: model.ProductStatInactive, Stock: 3}
		if canBePurchased(product, 1) {
			t.Fatal("purchase of inactive product was accepted")
		}
		if product.Stock != 3 {
			t.Errorf("stock after rejecting inactive product purchase = %d, want 3", product.Stock)
		}
	})

	t.Run("additional rule: successful purchase can reduce stock to zero but not below", func(t *testing.T) {
		product := model.Product{Status: model.ProductStatActive, Stock: 3}
		if !canBePurchased(product, 3) {
			t.Fatal("purchase of quantity equal to stock was rejected")
		}
		updatedProduct, err := decreasedStock(&product, 3)
		if err != nil {
			t.Fatalf("decreasedStock() error = %v, want nil", err)
		}
		if updatedProduct.Stock != 0 || product.Stock != 0 {
			t.Errorf("stock after successful purchase: returned=%d product=%d, want both 0",
				updatedProduct.Stock, product.Stock)
		}
	})
}
