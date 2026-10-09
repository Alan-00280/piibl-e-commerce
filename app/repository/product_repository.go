package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductBasePrice struct {
	Product   model.Product
	BasePrice int64
}

type ProductRepository interface {
	// Product
	FindByID(ctx context.Context, id int) (ProductBasePrice, error)
	FindAll(ctx context.Context, q model.CursorQuery) ([]ProductBasePrice, error)
	FindTenantIDByProductID(ctx context.Context, productID int) (int, error)
	// FindByStoreID includes inactive products; callers must authorize the tenant and store ownership.
	FindByStoreID(ctx context.Context, storeID int) ([]ProductBasePrice, error)
	Create(ctx context.Context, p model.Product, basePrice int64) (ProductBasePrice, error)
	Update(ctx context.Context, p model.Product) (model.Product, error)
	Delete(ctx context.Context, id int) error
	Deactivate(ctx context.Context, id int) error
	UpdateStock(ctx context.Context, id int, stock int) error
	UpdateStatus(ctx context.Context, id int, status model.ProductStat) error
	DecreaseStock(ctx context.Context, id int, quantity int) error

	// Variant Spaces
	CreateVariantSpace(ctx context.Context, vs model.ProductVariantSpace) (model.ProductVariantSpace, error)
	FindVariantSpaceByID(ctx context.Context, id int) (model.ProductVariantSpace, error)
	FindVariantSpacesByProductID(ctx context.Context, productID int) (model.ProductVariantSpaces, error)
	UpdateVariantSpace(ctx context.Context, vs model.ProductVariantSpace) (model.ProductVariantSpace, error)
	DeleteVariantSpace(ctx context.Context, id int) error

	// Product Variants
	CreateVariant(ctx context.Context, pv model.ProductVariant) (model.ProductVariant, error)
	FindVariantByID(ctx context.Context, id int) (model.ProductVariant, error)
	FindVariantsByProductID(ctx context.Context, productID int) ([]model.ProductVariant, error)
	FindVariantsByIDs(ctx context.Context, productID int, variantIDs []int) ([]model.ProductVariant, error)
	FindDefaultVariant(ctx context.Context, productID int) (model.ProductVariant, error)
	UpdateVariant(ctx context.Context, pv model.ProductVariant) (model.ProductVariant, error)
	DeleteVariant(ctx context.Context, id int) error
}

type productPostgresRepository struct {
	pool *pgxpool.Pool
}

var productColumns string = "p.id, p.store_id, p.category_id, p.name, COALESCE(p.description, ''), p.stock, p.status, p.created_at"

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &productPostgresRepository{
		pool: pool,
	}
}

// =========================================================
// PRODUCT METHODS
// =========================================================

func (r *productPostgresRepository) FindByID(ctx context.Context, id int) (ProductBasePrice, error) {
	result := ProductBasePrice{}

	query := fmt.Sprintf(`SELECT %s,
		COALESCE((SELECT pv.price
			FROM product_variants pv
			WHERE pv.product_id = p.id AND pv.name = '_default' AND pv.space_id IS NULL
			ORDER BY pv.id LIMIT 1), 0) AS base_price,
		COALESCE((SELECT AVG(pr.rating)
			FROM product_reviews pr
			WHERE pr.product_id = p.id), 0) AS overall_rating
		FROM products p WHERE p.id = $1`, productColumns)

	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.Product.ID, &result.Product.StoreID, &result.Product.CategoryID, &result.Product.Name,
		&result.Product.Description, &result.Product.Stock, &result.Product.Status, &result.Product.CreatedAt,
		&result.BasePrice, &result.Product.OverallRating,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProductBasePrice{}, ErrNotFound
		}
		if result.Product.Status == model.ProductStatInactive {
			return ProductBasePrice{}, ErrInactive
		}
		return ProductBasePrice{}, fmt.Errorf("[ERROR] can't get from products: %w", err)
	}

	return result, nil
}

func (r *productPostgresRepository) FindTenantIDByProductID(ctx context.Context, productID int) (int, error) {
	var tenantID int
	query := `SELECT s.tenant_id
		FROM products p
		JOIN stores s ON s.id = p.store_id
		WHERE p.id = $1`

	if err := r.pool.QueryRow(ctx, query, productID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("[ERROR] can't get tenant id by product id: %w", err)
	}

	return tenantID, nil
}

func (r *productPostgresRepository) FindAll(ctx context.Context, q model.CursorQuery) ([]ProductBasePrice, error) {
	where, args := buildFilterProduct(q)

	limit := q.Limit
	if limit < 1 {
		limit = 10
	}
	args = append(args, limit+1)

	sqlText := fmt.Sprintf(
		`SELECT %s,
			COALESCE((SELECT pv.price
				FROM product_variants pv
				WHERE pv.product_id = p.id AND pv.name = '_default' AND pv.space_id IS NULL
				ORDER BY pv.id LIMIT 1), 0) AS base_price,
			COALESCE((SELECT AVG(pr.rating)
				FROM product_reviews pr
				WHERE pr.product_id = p.id), 0) AS overall_rating
			FROM products p %s ORDER BY p.created_at DESC, p.id DESC LIMIT $%d`,
		productColumns, where, len(args),
	)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] can't get rows from products: %w", err)
	}
	defer rows.Close()

	result := []ProductBasePrice{}
	for rows.Next() {
		p, err := scanProductBasePrice(rows)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] can't scan row from products: %w", err)
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[ERROR] error query from products: %w", err)
	}

	return result, nil
}

func (r *productPostgresRepository) FindByStoreID(ctx context.Context, storeID int) ([]ProductBasePrice, error) {
	query := fmt.Sprintf(`SELECT %s,
		COALESCE((SELECT pv.price
			FROM product_variants pv
			WHERE pv.product_id = p.id AND pv.name = '_default' AND pv.space_id IS NULL
			ORDER BY pv.id LIMIT 1), 0) AS base_price,
		COALESCE((SELECT AVG(pr.rating)
			FROM product_reviews pr
			WHERE pr.product_id = p.id), 0) AS overall_rating
		FROM products p
		WHERE p.store_id = $1
		ORDER BY p.id ASC`, productColumns)

	rows, err := r.pool.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] can't get products by store id: %w", err)
	}
	defer rows.Close()

	result := []ProductBasePrice{}
	for rows.Next() {
		product, err := scanProductBasePrice(rows)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] can't scan product by store id: %w", err)
		}
		result = append(result, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[ERROR] error reading products by store id: %w", err)
	}

	return result, nil
}

func (r *productPostgresRepository) Create(ctx context.Context, p model.Product, basePrice int64) (ProductBasePrice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ProductBasePrice{}, fmt.Errorf("[ERROR] can't begin product creation transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO products (store_id, category_id, name, description, stock, status) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, created_at`

	if err := tx.QueryRow(ctx, query, p.StoreID, p.CategoryID, p.Name, p.Description, p.Stock, p.Status).Scan(
		&p.ID, &p.CreatedAt,
	); err != nil {
		return ProductBasePrice{}, fmt.Errorf("[ERROR] can't create product: %w", err)
	}

	defaultVariantQuery := `INSERT INTO product_variants (product_id, space_id, name, price)
	                        VALUES ($1, NULL, '_default', $2)`
	if _, err := tx.Exec(ctx, defaultVariantQuery, p.ID, basePrice); err != nil {
		return ProductBasePrice{}, fmt.Errorf("[ERROR] can't create default product variant: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return ProductBasePrice{}, fmt.Errorf("[ERROR] can't commit product creation: %w", err)
	}

	return ProductBasePrice{
		Product:   p,
		BasePrice: basePrice,
	}, nil
}

func (r *productPostgresRepository) Update(ctx context.Context, p model.Product) (model.Product, error) {
	query := `UPDATE products 
	          SET category_id = $1, name = $2, description = $3, stock = $4, status = $5, updated_at = NOW() 
	          WHERE id = $6 
	          RETURNING id, store_id, category_id, name, COALESCE(description, ''), stock, status, created_at`

	if err := r.pool.QueryRow(ctx, query, p.CategoryID, p.Name, p.Description, p.Stock, p.Status, p.ID).Scan(
		&p.ID, &p.StoreID, &p.CategoryID, &p.Name, &p.Description, &p.Stock, &p.Status, &p.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, ErrNotFound
		}
		return model.Product{}, fmt.Errorf("[ERROR] can't update product: %w", err)
	}

	return p, nil
}

func (r *productPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't delete product: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *productPostgresRepository) Deactivate(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(
		ctx,
		"UPDATE products SET status = $1, updated_at = NOW() WHERE id = $2",
		model.ProductStatInactive,
		id,
	)
	if err != nil {
		return fmt.Errorf("[ERROR] can't deactivate product: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *productPostgresRepository) UpdateStock(ctx context.Context, id int, stock int) error {

	tag, err := r.pool.Exec(ctx, "UPDATE products SET stock = $1, updated_at = NOW() WHERE id = $2", stock, id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't update product stock: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *productPostgresRepository) UpdateStatus(ctx context.Context, id int, status model.ProductStat) error {
	tag, err := r.pool.Exec(ctx, "UPDATE products SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't update product status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// DecreaseStock mengurangi stok secara aman
// Mencegah race condition dengan filter stock >= $1
func (r *productPostgresRepository) DecreaseStock(ctx context.Context, id int, quantity int) error {
	tag, err := r.pool.Exec(ctx, `UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2 AND stock >= $1 RETURNING stock`, quantity, id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't decrease stock: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("stok tidak mencukupi atau produk tidak ditemukan")
	}

	return nil
}

// =========================================================
// HELPER FUNCTIONS
// =========================================================

func buildFilterProduct(q model.CursorQuery) (string, []any) {
	where := " WHERE p.status = $1 "
	args := []any{model.ProductStatActive}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (p.name ILIKE $%d OR p.description ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if filter := q.ProductFilter; filter != nil {
		if filter.StoreID != nil && *filter.StoreID > 0 {
			where += fmt.Sprintf(" AND p.store_id = $%d", len(args)+1)
			args = append(args, *filter.StoreID)
		}
		if filter.CategoryID != nil && *filter.CategoryID > 0 {
			where += fmt.Sprintf(" AND p.category_id = $%d", len(args)+1)
			args = append(args, *filter.CategoryID)
		}
		if filter.PriceMin != nil {
			where += fmt.Sprintf(` AND EXISTS (
				SELECT 1 FROM product_variants pv
				WHERE pv.product_id = p.id
				  AND pv.name = '_default'
				  AND pv.space_id IS NULL
				  AND pv.price >= $%d
			)`, len(args)+1)
			args = append(args, *filter.PriceMin)
		}
		if filter.PriceMax != nil {
			where += fmt.Sprintf(` AND EXISTS (
				SELECT 1 FROM product_variants pv
				WHERE pv.product_id = p.id
				  AND pv.name = '_default'
				  AND pv.space_id IS NULL
				  AND pv.price <= $%d
			)`, len(args)+1)
			args = append(args, *filter.PriceMax)
		}
		if filter.MinRating != nil {
			where += fmt.Sprintf(` AND (
				SELECT AVG(pr.rating) FROM product_reviews pr
				WHERE pr.product_id = p.id
			) >= $%d`, len(args)+1)
			args = append(args, *filter.MinRating)
		}
		if filter.MaxRating != nil {
			where += fmt.Sprintf(` AND (
				SELECT AVG(pr.rating) FROM product_reviews pr
				WHERE pr.product_id = p.id
			) <= $%d`, len(args)+1)
			args = append(args, *filter.MaxRating)
		}
	}

	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (p.created_at, p.id) < ($%d, $%d)", len(args)-1, len(args))
	}

	return where, args
}

func scanProductBasePrice(row pgx.Row) (ProductBasePrice, error) {
	var result ProductBasePrice

	if err := row.Scan(
		&result.Product.ID,
		&result.Product.StoreID,
		&result.Product.CategoryID,
		&result.Product.Name,
		&result.Product.Description,
		&result.Product.Stock,
		&result.Product.Status,
		&result.Product.CreatedAt,
		&result.BasePrice,
		&result.Product.OverallRating,
	); err != nil {
		return ProductBasePrice{}, err
	}

	return result, nil
}
