package helper

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/gofiber/fiber/v2"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

// Negotiate memilih format response berdasarkan header Accept.
//
// Perbedaan yang wajib jelas:
// - Content-Type menjelaskan format yang SEDANG DIKIRIM pengirim.
// - Accept menjelaskan format yang DIINGINKAN penerima sebagai balasan.
func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))

	if accept == "" {
		return offered[0], nil
	}

	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable("format yang diminta tidak tersedia, pilih salah satu dari: " + strings.Join(offered, ", "))
	}

	return chosen, nil
}

// WriteOrdersCSV menuliskan order dan item order sebagai CSV, satu baris per item.
//
// Header Content-Disposition membuat browser menawarkan unduhan alih-alih
// menampilkan isinya sebagai teks mentah.
func WriteOrdersCSV(c *fiber.Ctx, orders []model.Order) error {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{
		"order_id",
		"order_group_id",
		"customer_id",
		"store_id",
		"status",
		"total",
		"order_created_at",
		"order_item_id",
		"product_id",
		"product_name",
		"variant",
		"price_at_purchase",
		"quantity",
		"subtotal",
	}

	if err := writer.Write(header); err != nil {
		return Internal(err)
	}

	for _, order := range orders {
		if len(order.OrderItems) == 0 {
			row := orderCSVRow(order, model.OrderItem{})
			if err := writer.Write(row); err != nil {
				return Internal(err)
			}
			continue
		}

		for _, item := range order.OrderItems {
			if err := writer.Write(orderCSVRow(order, item)); err != nil {
				return Internal(err)
			}
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return Internal(err)
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="orders.csv"`)

	return c.SendString(buffer.String())
}

func orderCSVRow(order model.Order, item model.OrderItem) []string {
	return []string{
		strconv.Itoa(order.ID),
		strconv.Itoa(order.OrderGroupID),
		strconv.Itoa(order.CustomerID),
		strconv.Itoa(order.StoreID),
		string(order.Status),
		strconv.FormatInt(order.Total, 10),
		order.CreatedAt.Format(time.RFC3339Nano),
		strconv.Itoa(item.ID),
		strconv.Itoa(item.ProductID),
		item.ProductName,
		item.VariantText,
		strconv.FormatInt(item.PriceAtPurchase, 10),
		strconv.Itoa(item.Quantity),
		strconv.FormatInt(item.Subtotal, 10),
	}
}
