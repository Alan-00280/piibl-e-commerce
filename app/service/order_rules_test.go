package service

import (
	"testing"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
)

func TestCalcOrderItemSubtotal(t *testing.T) {
	firstPositiveAdjustment := int64(5000)
	secondPositiveAdjustment := int64(3000)
	negativeAdjustment := int64(-2000)

	tests := []struct {
		name                 string
		defaultPrice         int64
		productVariantChosen []model.ProductVariant
		want                 int64
	}{
		{
			name:         "no selected variants returns default price",
			defaultPrice: 50000,
			want:         50000,
		},
		{
			name:         "one positive adjustment is added to default price",
			defaultPrice: 50000,
			productVariantChosen: []model.ProductVariant{
				{Name: "Extra shot", PriceAdjusment: &firstPositiveAdjustment},
			},
			want: 55000,
		},
		{
			name:         "multiple positive adjustments are added to default price",
			defaultPrice: 50000,
			productVariantChosen: []model.ProductVariant{
				{Name: "Extra shot", PriceAdjusment: &firstPositiveAdjustment},
				{Name: "Large size", PriceAdjusment: &secondPositiveAdjustment},
			},
			want: 58000,
		},
		{
			name:         "negative adjustment is added to default price",
			defaultPrice: 50000,
			productVariantChosen: []model.ProductVariant{
				{Name: "Discount", PriceAdjusment: &negativeAdjustment},
			},
			want: 48000,
		},
		{
			name:         "positive and negative adjustments are summed",
			defaultPrice: 50000,
			productVariantChosen: []model.ProductVariant{
				{Name: "Extra shot", PriceAdjusment: &firstPositiveAdjustment},
				{Name: "Discount", PriceAdjusment: &negativeAdjustment},
			},
			want: 53000,
		},
		{
			name:         "variant with nil adjustment contributes zero",
			defaultPrice: 50000,
			productVariantChosen: []model.ProductVariant{
				{Name: "Standard"},
				{Name: "Extra shot", PriceAdjusment: &firstPositiveAdjustment},
			},
			want: 55000,
		},
		{
			name:         "all nil adjustments preserve default price",
			defaultPrice: 50000,
			productVariantChosen: []model.ProductVariant{
				{Name: "Standard"},
				{Name: "Medium"},
			},
			want: 50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcOrderItemSubtotal(tt.defaultPrice, tt.productVariantChosen); got != tt.want {
				t.Errorf("calcOrderItemSubtotal(%d, %v) = %d, want %d",
					tt.defaultPrice, tt.productVariantChosen, got, tt.want)
			}
		})
	}
}

func TestCreateChoosenVariantText(t *testing.T) {
	colorSpaceID := 10
	sizeSpaceID := 20
	productTypeSpaceID := 30
	choosenVariants := []model.ProductVariant{
		{Name: "Merah Muda", VariantSpaceID: &colorSpaceID},
		{Name: "Extra Large", VariantSpaceID: &sizeSpaceID},
		{Name: "Besi Baja", VariantSpaceID: &productTypeSpaceID},
	}
	choosenSpaces := []model.ProductVariantSpace{
		{ID: colorSpaceID, Name: "Warna"},
		{ID: sizeSpaceID, Name: "Ukuran"},
		{ID: productTypeSpaceID, Name: "Tipe Produk"},
	}

	got := createChoosenVariantText(choosenVariants, choosenSpaces)
	want := "WARNA:Merah-Muda|UKURAN:Extra-Large|TIPE-PRODUK:Besi-Baja"
	if got != want {
		t.Errorf("createChoosenVariantText() = %q, want %q", got, want)
	}
}

func TestCheckChoosenVariants(t *testing.T) {
	colorSpaceID := 10
	sizeSpaceID := 20
	foreignSpaceID := 30

	colorRed := model.ProductVariant{Name: "Merah", VariantSpaceID: &colorSpaceID}
	colorBlue := model.ProductVariant{Name: "Biru", VariantSpaceID: &colorSpaceID}
	sizeMedium := model.ProductVariant{Name: "M", VariantSpaceID: &sizeSpaceID}
	foreignVariant := model.ProductVariant{Name: "Lain", VariantSpaceID: &foreignSpaceID}
	defaultVariant := model.ProductVariant{Name: "_default"}

	optionalSpaces := model.ProductVariantSpaces{
		ProductID: 1,
		VariantSpaces: []model.ProductVariantSpace{
			{ID: colorSpaceID, ProductID: 1, Name: "Warna"},
			{ID: sizeSpaceID, ProductID: 1, Name: "Ukuran"},
		},
	}
	mandatoryColorSpaces := model.ProductVariantSpaces{
		ProductID: 1,
		VariantSpaces: []model.ProductVariantSpace{
			{ID: colorSpaceID, ProductID: 1, Name: "Warna", IsMandatory: true},
			{ID: sizeSpaceID, ProductID: 1, Name: "Ukuran"},
		},
	}

	tests := []struct {
		name                 string
		productVariantSpaces model.ProductVariantSpaces
		choosens             []model.ProductVariant
		want                 string
	}{
		{
			name:                 "no mandatory spaces and no selections is valid",
			productVariantSpaces: optionalSpaces,
		},
		{
			name:                 "mandatory space with no selections is invalid",
			productVariantSpaces: mandatoryColorSpaces,
			want:                 "belum ada ruang yang dipilih",
		},
		{
			name:                 "selecting the mandatory space is valid",
			productVariantSpaces: mandatoryColorSpaces,
			choosens:             []model.ProductVariant{colorRed},
		},
		{
			name:                 "missing mandatory space is invalid",
			productVariantSpaces: mandatoryColorSpaces,
			choosens:             []model.ProductVariant{sizeMedium},
			want:                 "terdapat ruang yang belum dipilih",
		},
		{
			name:                 "one variant from each space is valid",
			productVariantSpaces: mandatoryColorSpaces,
			choosens:             []model.ProductVariant{colorRed, sizeMedium},
		},
		{
			name:                 "optional space can be selected without mandatory spaces",
			productVariantSpaces: optionalSpaces,
			choosens:             []model.ProductVariant{sizeMedium},
		},
		{
			name:                 "variant from a foreign space is invalid",
			productVariantSpaces: optionalSpaces,
			choosens:             []model.ProductVariant{foreignVariant},
			want:                 "variant bukan milik produk",
		},
		{
			name:                 "same variant chosen twice is invalid",
			productVariantSpaces: optionalSpaces,
			choosens:             []model.ProductVariant{colorRed, colorRed},
			want:                 "ruang varian Warna memiliki nilai ganda",
		},
		{
			name:                 "different variants from same space are invalid",
			productVariantSpaces: optionalSpaces,
			choosens:             []model.ProductVariant{colorRed, colorBlue},
			want:                 "ruang varian Warna memiliki nilai ganda",
		},
		{
			name:                 "explicit default variant is invalid",
			productVariantSpaces: optionalSpaces,
			choosens:             []model.ProductVariant{defaultVariant},
			want:                 "_default dilarang masuk ke choosen variant!",
		},
		{
			name:                 "normal variant without a space is invalid",
			productVariantSpaces: optionalSpaces,
			choosens: []model.ProductVariant{
				{Name: "Tanpa ruang"},
			},
			want: "_default dilarang masuk ke choosen variant!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkChoosenVariants(tt.productVariantSpaces, tt.choosens); got != tt.want {
				t.Errorf("checkChoosenVariants() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDuplicateChoosenVariants(t *testing.T) {
	firstSpaceID := 10
	secondSpaceID := 20

	tests := []struct {
		name                 string
		choosens             []model.ProductVariant
		wantDuplicateSpaceID int
		wantDuplicate        bool
	}{
		{
			name:                 "empty choices are not duplicate",
			wantDuplicateSpaceID: -1,
		},
		{
			name:                 "single variant is not duplicate",
			choosens:             []model.ProductVariant{{Name: "Merah", VariantSpaceID: &firstSpaceID}},
			wantDuplicateSpaceID: -1,
		},
		{
			wantDuplicateSpaceID: -1,
			name:                 "variants from different spaces are not duplicate",
			choosens: []model.ProductVariant{
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
				{Name: "M", VariantSpaceID: &secondSpaceID},
			},
		},
		{
			name: "different variants from same space are duplicate",
			choosens: []model.ProductVariant{
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
				{Name: "Biru", VariantSpaceID: &firstSpaceID},
			},
			wantDuplicateSpaceID: firstSpaceID,
			wantDuplicate:        true,
		},
		{
			name: "same variant chosen twice is duplicate",
			choosens: []model.ProductVariant{
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
			},
			wantDuplicateSpaceID: firstSpaceID,
			wantDuplicate:        true,
		},
		{
			wantDuplicateSpaceID: -2,
			name:                 "default and normal variant are duplicate spaces",
			choosens: []model.ProductVariant{
				{Name: "_default"},
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
			},
			wantDuplicate: true,
		},
		{
			name: "duplicate become true when exist an _default variant",
			choosens: []model.ProductVariant{
				{Name: "_default"},
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
				{Name: "Biru", VariantSpaceID: &firstSpaceID},
			},
			wantDuplicateSpaceID: -2,
			wantDuplicate:        true,
		},
		{
			name: "repeated space among more than two variants is duplicate",
			choosens: []model.ProductVariant{
				{Name: "Merah", VariantSpaceID: &firstSpaceID},
				{Name: "M", VariantSpaceID: &secondSpaceID},
				{Name: "Biru", VariantSpaceID: &firstSpaceID},
			},
			wantDuplicateSpaceID: firstSpaceID,
			wantDuplicate:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSpaceID, gotDuplicate := duplicateChoosenVariants(tt.choosens)
			if gotDuplicate != tt.wantDuplicate {
				t.Errorf("duplicateChoosenVariants() duplicate = %t, want %t", gotDuplicate, tt.wantDuplicate)
			}
			if gotSpaceID != tt.wantDuplicateSpaceID {
				t.Errorf("duplicateChoosenVariants() space ID = %d, want %d", gotSpaceID, tt.wantDuplicateSpaceID)
			}
		})
	}
}

func TestDuplicateChoosenVariantsWithDefaultBetweenDuplicates(t *testing.T) {
	spaceID := 10
	choosens := []model.ProductVariant{
		{Name: "Merah", VariantSpaceID: &spaceID},
		{Name: "_default"},
		{Name: "Biru", VariantSpaceID: &spaceID},
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Errorf("duplicateChoosenVariants() panicked with _default between duplicate variants: %v", recovered)
		}
	}()

	gotSpaceID, gotDuplicate := duplicateChoosenVariants(choosens)
	if !gotDuplicate {
		t.Fatal("duplicateChoosenVariants() duplicate = false, want true")
	}
	if gotSpaceID != -2 {
		t.Errorf("duplicateChoosenVariants() space ID = %d, want %d", gotSpaceID, -2)
	}
}

func TestGroupItemPerStore(t *testing.T) {
	items := []*model.ProductToBeGrouped{
		{
			Product:         model.Product{ID: 1, StoreID: 10, Name: "Coffee"},
			VariantText:     "SIZE:Large",
			PriceAtPurchase: 25000,
			Quantity:        2,
		},
		{
			Product:         model.Product{ID: 2, StoreID: 10, Name: "Tea"},
			PriceAtPurchase: 15000,
			Quantity:        1,
		},
		{
			Product:         model.Product{ID: 3, StoreID: 20, Name: "Bread"},
			PriceAtPurchase: 10000,
			Quantity:        1,
		},
	}

	const customerID = 7
	orders := groupItemPerStore(items, customerID)
	if len(orders) != 2 {
		t.Fatalf("groupItemPerStore() returned %d orders, want 2", len(orders))
	}

	orderItemsPerStore := make(map[int][]model.OrderItem)
	for _, order := range orders {
		if order.CustomerID != customerID {
			t.Errorf("order customer ID = %d, want %d", order.CustomerID, customerID)
		}
		if order.Status != model.OrderStatusCreated {
			t.Errorf("order status = %q, want %q", order.Status, model.OrderStatusCreated)
		}
		orderItemsPerStore[order.StoreID] = order.OrderItems
	}

	store10Items := orderItemsPerStore[10]
	if len(store10Items) != 2 {
		t.Fatalf("store 10 has %d order items, want 2", len(store10Items))
	}
	if got := store10Items[0]; got.ProductID != 1 ||
		got.ProductName != "Coffee" ||
		got.VariantText != "SIZE:Large" ||
		got.PriceAtPurchase != 25000 ||
		got.Quantity != 2 ||
		got.Subtotal != 0 {
		t.Errorf("first store 10 order item = %+v, want the Coffee purchase details", got)
	}

	store20Items := orderItemsPerStore[20]
	if len(store20Items) != 1 {
		t.Fatalf("store 20 has %d order items, want 1", len(store20Items))
	}
	if got := store20Items[0]; got.ProductID != 3 ||
		got.ProductName != "Bread" ||
		got.PriceAtPurchase != 10000 ||
		got.Quantity != 1 ||
		got.Subtotal != 0 {
		t.Errorf("first store 20 order item = %+v, want the Bread purchase details", got)
	}
}

func TestCountTotalSubtotalSinglePurchase(t *testing.T) {
	orderItems := []*model.OrderItem{
		{
			ProductID:       1,
			ProductName:     "Coffee",
			PriceAtPurchase: 25000,
			Quantity:        2,
		},
	}
	order := model.Order{}

	appliedItems, result := countTotalSubtotal(orderItems, &order)
	if len(appliedItems) != 1 {
		t.Fatalf("countTotalSubtotal() returned %d items, want 1", len(appliedItems))
	}
	if got, want := appliedItems[0].Subtotal, int64(50000); got != want {
		t.Errorf("item subtotal = %d, want %d", got, want)
	}
	if got, want := result.Total, int64(50000); got != want {
		t.Errorf("order total = %d, want %d", got, want)
	}
}
