package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	// Product
	FindByID(ctx context.Context, id int) (model.Product, error)
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Product, int, error)
	Create(ctx context.Context, p model.Product) (model.Product, error)
	Update(ctx context.Context, p model.Product) (model.Product, error)
	Delete(ctx context.Context, id int) error
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

var productColumns string = "id, store_id, category_id, name, COALESCE(description, ''), stock, status, created_at"

var productSortColumn = map[string]string{
	"id":         "id",
	"name":       "name",
	"stock":      "stock",
	"created_at": "created_at",
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &productPostgresRepository{
		pool: pool,
	}
}

// =========================================================
// PRODUCT METHODS
// =========================================================

func (r *productPostgresRepository) FindByID(ctx context.Context, id int) (model.Product, error) {
	result := model.Product{}

	query := fmt.Sprintf("SELECT %s FROM products WHERE id = $1", productColumns)

	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.StoreID, &result.CategoryID, &result.Name,
		&result.Description, &result.Stock, &result.Status, &result.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, ErrNotFound
		}
		return model.Product{}, fmt.Errorf("[ERROR] can't get from products: %w", err)
	}

	return result, nil
}

func (r *productPostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Product, int, error) {
	where, args := buildFilterProduct(q)

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM products"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("[ERROR] count total products: %w", err)
	}

	direction := "ASC"
	if q.Order == "desc" {
		direction = "DESC"
	}

	sortCol := "id"
	if col, ok := productSortColumn[q.Sort]; ok {
		sortCol = col
	}

	sqlText := fmt.Sprintf(
		`SELECT %s FROM products %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		productColumns, where, sortCol, direction, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("[ERROR] can't get rows from products: %w", err)
	}
	defer rows.Close()

	result := []model.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("[ERROR] can't scan row from products: %w", err)
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("[ERROR] error query from products: %w", err)
	}

	return result, total, nil
}

func (r *productPostgresRepository) Create(ctx context.Context, p model.Product) (model.Product, error) {
	query := `INSERT INTO products (store_id, category_id, name, description, stock, status) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, created_at`

	if err := r.pool.QueryRow(ctx, query, p.StoreID, p.CategoryID, p.Name, p.Description, p.Stock, p.Status).Scan(
		&p.ID, &p.CreatedAt,
	); err != nil {
		return model.Product{}, fmt.Errorf("[ERROR] can't create product: %w", err)
	}

	return p, nil
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
	tag, err := r.pool.Exec(ctx, `UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2 AND stock >= $1`, quantity, id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't decrease stock: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("stok tidak mencukupi atau produk tidak ditemukan")
	}

	return nil
}

// =========================================================
// VARIANT SPACES METHODS
// =========================================================

func (r *productPostgresRepository) CreateVariantSpace(ctx context.Context, vs model.ProductVariantSpace) (model.ProductVariantSpace, error) {
	query := `INSERT INTO variant_spaces (product_id, name, is_mandatory) VALUES ($1, $2, $3) RETURNING id`

	if err := r.pool.QueryRow(ctx, query, vs.ProductID, vs.Name, vs.IsMandatory).Scan(&vs.ID); err != nil {
		if isUniqueViolation(err) {
			return model.ProductVariantSpace{}, ErrDuplicate
		}
		return model.ProductVariantSpace{}, fmt.Errorf("[ERROR] can't create variant space: %w", err)
	}

	return vs, nil
}

func (r *productPostgresRepository) FindVariantSpaceByID(ctx context.Context, id int) (model.ProductVariantSpace, error) {
	var vs model.ProductVariantSpace
	query := `SELECT id, product_id, name, is_mandatory FROM variant_spaces WHERE id = $1`

	if err := r.pool.QueryRow(ctx, query, id).Scan(&vs.ID, &vs.ProductID, &vs.Name, &vs.IsMandatory); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ProductVariantSpace{}, ErrNotFound
		}
		return model.ProductVariantSpace{}, fmt.Errorf("[ERROR] can't get variant space: %w", err)
	}

	return vs, nil
}

func (r *productPostgresRepository) FindVariantSpacesByProductID(ctx context.Context, productID int) (model.ProductVariantSpaces, error) {
	result := model.ProductVariantSpaces{
		ProductID:     productID,
		VariantSpaces: []model.ProductVariantSpace{},
	}

	spaceQuery := `SELECT id, product_id, name, is_mandatory FROM variant_spaces WHERE product_id = $1 ORDER BY id ASC`
	spaceRows, err := r.pool.Query(ctx, spaceQuery, productID)
	if err != nil {
		return result, fmt.Errorf("[ERROR] can't query variant spaces: %w", err)
	}
	defer spaceRows.Close()

	spaceMap := make(map[int]int)
	for spaceRows.Next() {
		var vs model.ProductVariantSpace
		if err := spaceRows.Scan(&vs.ID, &vs.ProductID, &vs.Name, &vs.IsMandatory); err != nil {
			return result, fmt.Errorf("[ERROR] can't scan variant space: %w", err)
		}
		vs.ProductVariants = []model.ProductVariant{}
		spaceMap[vs.ID] = len(result.VariantSpaces)
		result.VariantSpaces = append(result.VariantSpaces, vs)
	}
	if err := spaceRows.Err(); err != nil {
		return result, fmt.Errorf("[ERROR] error reading variant spaces: %w", err)
	}

	variantQuery := `SELECT id, product_id, space_id, name, price FROM product_variants WHERE product_id = $1 AND space_id IS NOT NULL ORDER BY id ASC`
	varRows, err := r.pool.Query(ctx, variantQuery, productID)
	if err != nil {
		return result, fmt.Errorf("[ERROR] can't query product variants: %w", err)
	}
	defer varRows.Close()

	for varRows.Next() {
		var pv model.ProductVariant
		if err := varRows.Scan(&pv.ID, &pv.ProductID, &pv.VariantSpaceID, &pv.Name, &pv.PriceAdjusment); err != nil {
			return result, fmt.Errorf("[ERROR] can't scan product variant: %w", err)
		}
		if pv.VariantSpaceID != nil {
			if idx, ok := spaceMap[*pv.VariantSpaceID]; ok {
				result.VariantSpaces[idx].ProductVariants = append(result.VariantSpaces[idx].ProductVariants, pv)
			}
		}
	}
	if err := varRows.Err(); err != nil {
		return result, fmt.Errorf("[ERROR] error reading product variants: %w", err)
	}

	return result, nil
}

func (r *productPostgresRepository) UpdateVariantSpace(ctx context.Context, vs model.ProductVariantSpace) (model.ProductVariantSpace, error) {
	query := `UPDATE variant_spaces SET name = $1, is_mandatory = $2 WHERE id = $3 RETURNING id, product_id, name, is_mandatory`

	if err := r.pool.QueryRow(ctx, query, vs.Name, vs.IsMandatory, vs.ID).Scan(
		&vs.ID, &vs.ProductID, &vs.Name, &vs.IsMandatory,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ProductVariantSpace{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.ProductVariantSpace{}, ErrDuplicate
		}
		return model.ProductVariantSpace{}, fmt.Errorf("[ERROR] can't update variant space: %w", err)
	}

	return vs, nil
}

func (r *productPostgresRepository) DeleteVariantSpace(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM variant_spaces WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't delete variant space: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// =========================================================
// PRODUCT VARIANTS METHODS
// =========================================================

func (r *productPostgresRepository) CreateVariant(ctx context.Context, pv model.ProductVariant) (model.ProductVariant, error) {
	query := `INSERT INTO product_variants (product_id, space_id, name, price) VALUES ($1, $2, $3, $4) RETURNING id`

	if err := r.pool.QueryRow(ctx, query, pv.ProductID, pv.VariantSpaceID, pv.Name, pv.PriceAdjusment).Scan(&pv.ID); err != nil {
		if isUniqueViolation(err) {
			return model.ProductVariant{}, ErrDuplicate
		}
		return model.ProductVariant{}, fmt.Errorf("[ERROR] can't create product variant: %w", err)
	}

	return pv, nil
}

func (r *productPostgresRepository) FindVariantByID(ctx context.Context, id int) (model.ProductVariant, error) {
	var pv model.ProductVariant
	query := `SELECT id, product_id, space_id, name, price FROM product_variants WHERE id = $1`

	if err := r.pool.QueryRow(ctx, query, id).Scan(&pv.ID, &pv.ProductID, &pv.VariantSpaceID, &pv.Name, &pv.PriceAdjusment); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ProductVariant{}, ErrNotFound
		}
		return model.ProductVariant{}, fmt.Errorf("[ERROR] can't get product variant: %w", err)
	}

	return pv, nil
}

func (r *productPostgresRepository) FindVariantsByProductID(ctx context.Context, productID int) ([]model.ProductVariant, error) {
	query := `SELECT id, product_id, space_id, name, price FROM product_variants WHERE product_id = $1 ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] can't query product variants: %w", err)
	}
	defer rows.Close()

	var variants []model.ProductVariant
	for rows.Next() {
		var v model.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.VariantSpaceID, &v.Name, &v.PriceAdjusment); err != nil {
			return nil, fmt.Errorf("[ERROR] can't scan product variant: %w", err)
		}
		variants = append(variants, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[ERROR] error reading product variants: %w", err)
	}

	return variants, nil
}

func (r *productPostgresRepository) FindVariantsByIDs(ctx context.Context, productID int, variantIDs []int) ([]model.ProductVariant, error) {
	if len(variantIDs) == 0 {
		return []model.ProductVariant{}, nil
	}

	query := `SELECT id, product_id, space_id, name, price 
	          FROM product_variants 
	          WHERE product_id = $1 AND id = ANY($2) 
	          ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, query, productID, variantIDs)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] can't query variants by ids: %w", err)
	}
	defer rows.Close()

	var result []model.ProductVariant
	for rows.Next() {
		var pv model.ProductVariant
		if err := rows.Scan(&pv.ID, &pv.ProductID, &pv.VariantSpaceID, &pv.Name, &pv.PriceAdjusment); err != nil {
			return nil, fmt.Errorf("[ERROR] can't scan variant: %w", err)
		}
		result = append(result, pv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[ERROR] error reading variants: %w", err)
	}

	return result, nil
}

func (r *productPostgresRepository) FindDefaultVariant(ctx context.Context, productID int) (model.ProductVariant, error) {
	var pv model.ProductVariant
	query := `SELECT id, product_id, space_id, name, price 
	          FROM product_variants 
	          WHERE product_id = $1 AND name = '_default' AND space_id IS NULL`

	if err := r.pool.QueryRow(ctx, query, productID).Scan(
		&pv.ID, &pv.ProductID, &pv.VariantSpaceID, &pv.Name, &pv.PriceAdjusment,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ProductVariant{}, ErrNotFound
		}
		return model.ProductVariant{}, fmt.Errorf("[ERROR] can't get default variant: %w", err)
	}

	return pv, nil
}

func (r *productPostgresRepository) UpdateVariant(ctx context.Context, pv model.ProductVariant) (model.ProductVariant, error) {
	query := `UPDATE product_variants SET name = $1, price = $2 WHERE id = $3 RETURNING id, product_id, space_id, name, price`

	if err := r.pool.QueryRow(ctx, query, pv.Name, pv.PriceAdjusment, pv.ID).Scan(
		&pv.ID, &pv.ProductID, &pv.VariantSpaceID, &pv.Name, &pv.PriceAdjusment,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ProductVariant{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.ProductVariant{}, ErrDuplicate
		}
		return model.ProductVariant{}, fmt.Errorf("[ERROR] can't update product variant: %w", err)
	}

	return pv, nil
}

func (r *productPostgresRepository) DeleteVariant(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM product_variants WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't delete product variant: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// =========================================================
// HELPER FUNCTIONS
// =========================================================

func buildFilterProduct(q model.ListQuery) (string, []any) {
	where := " WHERE 1=1 "
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.IsActive != nil {
		where += fmt.Sprintf(" AND status = $%d", len(args)+1)
		if *q.IsActive {
			args = append(args, model.ProductStatActive)
		} else {
			args = append(args, model.ProductStatInactive)
		}
	}

	if q.ProductFilter != nil {
		if q.ProductFilter.StoreID > 0 {
			where += fmt.Sprintf(" AND store_id = $%d", len(args)+1)
			args = append(args, q.ProductFilter.StoreID)
		}
		if q.ProductFilter.CategoryID > 0 {
			where += fmt.Sprintf(" AND category_id = $%d", len(args)+1)
			args = append(args, q.ProductFilter.CategoryID)
		}
	}

	return where, args
}

func scanProduct(rows pgx.Rows) (model.Product, error) {
	var p model.Product

	if err := rows.Scan(
		&p.ID, &p.StoreID, &p.CategoryID, &p.Name,
		&p.Description, &p.Stock, &p.Status, &p.CreatedAt,
	); err != nil {
		return model.Product{}, err
	}

	return p, nil
}
