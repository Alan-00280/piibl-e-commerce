package service

func checkVariantAdjustmentPrice(defaultPrice int64, adjustmentPrices []int64) string {
	var totalAdjustment int64
	totalAdjustment = 0

	for _, adjustmentPrice := range adjustmentPrices {
		totalAdjustment += adjustmentPrice
	}

	newPrice := totalAdjustment + defaultPrice
	if newPrice < 100 {
		return "Penyesuaian harga varian tidak boleh mengubah harga dasar hingga di bawah 100"
	}

	return ""
}
