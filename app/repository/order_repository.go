package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	FindAll(ctx context.Context, q model.ListQuery, currentUser model.AuthUser) ([]model.Order, int, error)
	FindByID(ctx context.Context, id int, currentUser model.AuthUser) (model.Order, error)
	UpdateTenantOrderStatus(ctx context.Context, id int, tenantID int, status model.OrderStat) (model.Order, error)
	FindCheckoutProducts(ctx context.Context, productIDs []int) (map[int]model.CheckoutProduct, error)
	FindCheckoutVariantSpaces(ctx context.Context, productIDs []int) (map[int]model.ProductVariantSpaces, error)
	FindCheckoutVariants(ctx context.Context, variantIDs []int) (map[int]model.ProductVariant, error)
	CreateCheckout(ctx context.Context, customerID int, checkoutItems []model.CheckoutItem, orders []model.Order) (model.CheckoutResponse, error)
}

type orderPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &orderPostgresRepository{pool: pool}
}

func (r *orderPostgresRepository) FindAll(
	ctx context.Context,
	q model.ListQuery,
	currentUser model.AuthUser,
) ([]model.Order, int, error) {
	from, where, args, err := buildFilterOrder(q, currentUser)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*)"+from+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	sortColumns := map[string]string{
		"id":         "o.id",
		"created_at": "o.created_at",
		"status":     "o.status",
		"total":      "o.total",
		"store_id":   "o.store_id",
	}
	sortColumn := sortColumns[q.Sort]
	if sortColumn == "" {
		sortColumn = sortColumns["id"]
	}

	direction := "ASC"
	if q.Order == "desc" {
		direction = "DESC"
	}

	query := fmt.Sprintf(
		`SELECT o.id, o.order_group_id, o.customer_id, o.store_id, o.status, o.total, o.created_at
		 %s%s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		from,
		where,
		sortColumn,
		direction,
		len(args)+1,
		len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0, q.Limit)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.OrderGroupID,
			&order.CustomerID,
			&order.StoreID,
			&order.Status,
			&order.Total,
			&order.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		order.OrderItems = []model.OrderItem{}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read orders: %w", err)
	}

	return orders, total, nil
}

func (r *orderPostgresRepository) FindByID(
	ctx context.Context,
	id int,
	currentUser model.AuthUser,
) (model.Order, error) {
	from, where, args, err := buildFilterOrder(model.ListQuery{}, currentUser)
	if err != nil {
		return model.Order{}, err
	}
	args = append(args, id)
	where += fmt.Sprintf(" AND o.id = $%d", len(args))

	var order model.Order
	if err := r.pool.QueryRow(ctx,
		`SELECT o.id, o.order_group_id, o.customer_id, o.store_id, o.status, o.total, o.created_at `+from+where,
		args...,
	).Scan(
		&order.ID,
		&order.OrderGroupID,
		&order.CustomerID,
		&order.StoreID,
		&order.Status,
		&order.Total,
		&order.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrNotFound
		}
		return model.Order{}, fmt.Errorf("find order by id: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_name, variant, price, quantity, subtotal
		 FROM order_items
		 WHERE order_id = $1
		 ORDER BY id`,
		order.ID,
	)
	if err != nil {
		return model.Order{}, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	order.OrderItems = make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.VariantText,
			&item.PriceAtPurchase,
			&item.Quantity,
			&item.Subtotal,
		); err != nil {
			return model.Order{}, fmt.Errorf("scan order item: %w", err)
		}
		order.OrderItems = append(order.OrderItems, item)
	}
	if err := rows.Err(); err != nil {
		return model.Order{}, fmt.Errorf("read order items: %w", err)
	}

	return order, nil
}

func (r *orderPostgresRepository) UpdateTenantOrderStatus(
	ctx context.Context,
	id int,
	tenantID int,
	status model.OrderStat,
) (model.Order, error) {
	var order model.Order
	err := r.pool.QueryRow(ctx,
		`UPDATE orders o
		 SET status = $1
		 FROM stores s
		 WHERE o.id = $2
		   AND o.store_id = s.id
		   AND s.tenant_id = $3
		   AND o.status = 'CREATED'
		 RETURNING o.id, o.order_group_id, o.customer_id, o.store_id, o.status, o.total, o.created_at`,
		status,
		id,
		tenantID,
	).Scan(
		&order.ID,
		&order.OrderGroupID,
		&order.CustomerID,
		&order.StoreID,
		&order.Status,
		&order.Total,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var exists bool
			if err := r.pool.QueryRow(ctx,
				`SELECT EXISTS (
					SELECT 1
					FROM orders o
					JOIN stores s ON s.id = o.store_id
					WHERE o.id = $1 AND s.tenant_id = $2
				)`,
				id,
				tenantID,
			).Scan(&exists); err != nil {
				return model.Order{}, fmt.Errorf("check tenant order after status update miss: %w", err)
			}
			if !exists {
				return model.Order{}, ErrNotFound
			}
			return model.Order{}, ErrOrderStatusConflict
		}
		return model.Order{}, fmt.Errorf("update tenant order status: %w", err)
	}

	order.OrderItems = []model.OrderItem{}
	return order, nil
}

func buildFilterOrder(q model.ListQuery, currentUser model.AuthUser) (string, string, []any, error) {
	var (
		from  string
		where string
	)
	args := []any{currentUser.UserID}

	switch currentUser.Role {
	case model.RoleCustomer:
		from = " FROM orders o"
		where = " WHERE o.customer_id = $1"
	case model.RoleTenant:
		from = " FROM orders o JOIN stores s ON s.id = o.store_id"
		where = " WHERE s.tenant_id = $1"
	default:
		return "", "", nil, fmt.Errorf("unsupported role for order listing: %s", currentUser.Role)
	}

	if q.OrderFilter != nil && q.OrderFilter.OrderStat != nil {
		where += fmt.Sprintf(" AND o.status = $%d", len(args)+1)
		args = append(args, *q.OrderFilter.OrderStat)
	}

	return from, where, args, nil
}

func (r *orderPostgresRepository) FindCheckoutProducts(
	ctx context.Context,
	productIDs []int,
) (map[int]model.CheckoutProduct, error) {
	ids := make([]int64, len(productIDs))
	for i, id := range productIDs {
		ids[i] = int64(id)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT p.id, p.store_id, p.category_id, p.name,
			COALESCE(p.description, ''), p.stock, p.status, p.created_at, pv.price
		 FROM products p
		 LEFT JOIN product_variants pv
			ON pv.product_id = p.id AND pv.name = '_default' AND pv.space_id IS NULL
		 WHERE p.id = ANY($1)
		 ORDER BY p.id`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("batch load checkout products: %w", err)
	}
	defer rows.Close()

	products := make(map[int]model.CheckoutProduct, len(productIDs))
	for rows.Next() {
		var product model.CheckoutProduct
		var basePrice *int64
		if err := rows.Scan(
			&product.Product.ID,
			&product.Product.StoreID,
			&product.Product.CategoryID,
			&product.Product.Name,
			&product.Product.Description,
			&product.Product.Stock,
			&product.Product.Status,
			&product.Product.CreatedAt,
			&basePrice,
		); err != nil {
			return nil, fmt.Errorf("scan checkout product: %w", err)
		}
		if basePrice != nil {
			product.BasePrice = *basePrice
			product.HasBasePrice = true
		}
		products[product.Product.ID] = product
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read checkout products: %w", err)
	}
	return products, nil
}

func (r *orderPostgresRepository) FindCheckoutVariantSpaces(
	ctx context.Context,
	productIDs []int,
) (map[int]model.ProductVariantSpaces, error) {
	ids := make([]int64, len(productIDs))
	result := make(map[int]model.ProductVariantSpaces, len(productIDs))
	for i, id := range productIDs {
		ids[i] = int64(id)
		result[id] = model.ProductVariantSpaces{
			ProductID:     id,
			VariantSpaces: []model.ProductVariantSpace{},
		}
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, name, is_mandatory
		 FROM variant_spaces
		 WHERE product_id = ANY($1)
		 ORDER BY product_id, id`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("batch load checkout variant spaces: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var space model.ProductVariantSpace
		if err := rows.Scan(&space.ID, &space.ProductID, &space.Name, &space.IsMandatory); err != nil {
			return nil, fmt.Errorf("scan checkout variant space: %w", err)
		}
		space.ProductVariants = []model.ProductVariant{}
		productSpaces := result[space.ProductID]
		productSpaces.VariantSpaces = append(productSpaces.VariantSpaces, space)
		result[space.ProductID] = productSpaces
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read checkout variant spaces: %w", err)
	}
	return result, nil
}

func (r *orderPostgresRepository) FindCheckoutVariants(
	ctx context.Context,
	variantIDs []int,
) (map[int]model.ProductVariant, error) {
	variants := make(map[int]model.ProductVariant, len(variantIDs))
	if len(variantIDs) == 0 {
		return variants, nil
	}

	ids := make([]int64, len(variantIDs))
	for i, id := range variantIDs {
		ids[i] = int64(id)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, space_id, name, price
		 FROM product_variants
		 WHERE id = ANY($1)
		 ORDER BY id`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("batch load checkout variants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var variant model.ProductVariant
		if err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.VariantSpaceID,
			&variant.Name,
			&variant.PriceAdjusment,
		); err != nil {
			return nil, fmt.Errorf("scan checkout variant: %w", err)
		}
		variants[variant.ID] = variant
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read checkout variants: %w", err)
	}
	return variants, nil
}

func (r *orderPostgresRepository) CreateCheckout(
	ctx context.Context,
	customerID int,
	checkoutItems []model.CheckoutItem,
	orders []model.Order,
) (model.CheckoutResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.CheckoutResponse{}, fmt.Errorf("begin checkout transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	quantitiesByProduct := make(map[int]int)
	for _, item := range checkoutItems {
		quantitiesByProduct[item.ProductID] += item.Quantity
	}
	productIDs := make([]int, 0, len(quantitiesByProduct))
	for productID := range quantitiesByProduct {
		productIDs = append(productIDs, productID)
	}
	sort.Ints(productIDs)

	for _, productID := range productIDs {
		tag, err := tx.Exec(ctx,
			`UPDATE products
			 SET stock = stock - $1, updated_at = NOW()
			 WHERE id = $2 AND stock >= $1 AND status = 'ACTIVE'`,
			quantitiesByProduct[productID], productID,
		)
		if err != nil {
			return model.CheckoutResponse{}, fmt.Errorf("conditionally decrease stock for product %d: %w", productID, err)
		}
		if tag.RowsAffected() == 0 {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				return model.CheckoutResponse{}, fmt.Errorf("rollback checkout after stock conflict: %w", err)
			}
			return model.CheckoutResponse{}, ErrStockConflict
		}
	}

	result := model.CheckoutResponse{
		Orders: make([]model.CheckoutOrderResponse, 0, len(orders)),
	}
	if err := tx.QueryRow(ctx,
		`INSERT INTO order_groups (customer_id, created_at)
		 VALUES ($1, NOW())
		 RETURNING id, created_at`,
		customerID,
	).Scan(&result.OrderGroupID, &result.CreatedAt); err != nil {
		return model.CheckoutResponse{}, fmt.Errorf("insert checkout order group: %w", err)
	}

	for orderIndex := range orders {
		order := &orders[orderIndex]
		if err := tx.QueryRow(ctx,
			`INSERT INTO orders (order_group_id, customer_id, store_id, status, total, created_at)
			 VALUES ($1, $2, $3, 'CREATED', $4, NOW())
			 RETURNING id, created_at`,
			result.OrderGroupID,
			customerID,
			order.StoreID,
			order.Total,
		).Scan(&order.ID, &order.CreatedAt); err != nil {
			return model.CheckoutResponse{}, fmt.Errorf("insert checkout order for store %d: %w", order.StoreID, err)
		}
		order.OrderGroupID = result.OrderGroupID
		order.CustomerID = customerID

		orderResponse := model.CheckoutOrderResponse{
			OrderID: order.ID,
			StoreID: order.StoreID,
			Status:  order.Status,
			Total:   order.Total,
			Items:   make([]model.CheckoutOrderItemResponse, 0, len(order.OrderItems)),
		}
		for itemIndex := range order.OrderItems {
			item := &order.OrderItems[itemIndex]
			if err := tx.QueryRow(ctx,
				`INSERT INTO order_items (order_id, product_id, product_name, variant, price, quantity, subtotal, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
				 RETURNING id`,
				order.ID,
				item.ProductID,
				item.ProductName,
				item.VariantText,
				item.PriceAtPurchase,
				item.Quantity,
				item.Subtotal,
			).Scan(&item.ID); err != nil {
				return model.CheckoutResponse{}, fmt.Errorf("insert order item for product %d: %w", item.ProductID, err)
			}
			item.OrderID = order.ID
			orderResponse.Items = append(orderResponse.Items, model.CheckoutOrderItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Variant:     item.VariantText,
				Price:       item.PriceAtPurchase,
				Quantity:    item.Quantity,
				Subtotal:    item.Subtotal,
			})
		}
		result.Orders = append(result.Orders, orderResponse)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.CheckoutResponse{}, fmt.Errorf("commit checkout transaction: %w", err)
	}
	return result, nil
}
