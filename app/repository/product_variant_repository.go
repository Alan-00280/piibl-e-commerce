package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
)

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
