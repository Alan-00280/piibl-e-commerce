package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
)

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
