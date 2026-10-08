package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoreRepository interface {
	FindByID(ctx context.Context, id int) (model.Stores, error)
	FindByOwnerID(ctx context.Context, ownerID int) (model.Stores, error)
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Stores, int, error)
	Create(ctx context.Context, s model.Stores) (model.Stores, error)
	Update(ctx context.Context, s model.Stores) (model.Stores, error)
	Delete(ctx context.Context, id int) error
}

type storePostgresRepository struct {
	pool *pgxpool.Pool
}

var storeColumns string = "id, tenant_id, name, COALESCE(description, ''), is_active, created_at"

var storeSortColumn = map[string]string{
	"id":         "id",
	"name":       "name",
	"created_at": "created_at",
}

func NewStoreRepository(pool *pgxpool.Pool) StoreRepository {
	return &storePostgresRepository{
		pool: pool,
	}
}

func (r *storePostgresRepository) FindByID(ctx context.Context, id int) (model.Stores, error) {
	result := model.Stores{}
	var createdAt time.Time

	query := fmt.Sprintf("SELECT %s FROM stores WHERE id = $1", storeColumns)

	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.OwnerID, &result.Name, &result.Description, &result.IsActive, &createdAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Stores{}, ErrNotFound
		}
		return model.Stores{}, fmt.Errorf("[ERROR] can't get from stores: %w", err)
	}

	result.CreatedAt = createdAt.Format(time.RFC3339)
	return result, nil
}

func (r *storePostgresRepository) FindByOwnerID(ctx context.Context, ownerID int) (model.Stores, error) {
	result := model.Stores{}
	var createdAt time.Time

	query := fmt.Sprintf("SELECT %s FROM stores WHERE tenant_id = $1", storeColumns)

	if err := r.pool.QueryRow(ctx, query, ownerID).Scan(
		&result.ID, &result.OwnerID, &result.Name, &result.Description, &result.IsActive, &createdAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Stores{}, ErrNotFound
		}
		return model.Stores{}, fmt.Errorf("[ERROR] can't get store by owner id: %w", err)
	}

	result.CreatedAt = createdAt.Format(time.RFC3339)
	return result, nil
}

func (r *storePostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Stores, int, error) {
	where, args := buildFilterStore(q)

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM stores"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("[ERROR] count total stores: %w", err)
	}

	direction := "ASC"
	if q.Order == "desc" {
		direction = "DESC"
	}

	sortCol := "id"
	if col, ok := storeSortColumn[q.Sort]; ok {
		sortCol = col
	}

	sqlText := fmt.Sprintf(
		`SELECT %s FROM stores %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		storeColumns, where, sortCol, direction, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("[ERROR] can't get rows from stores: %w", err)
	}
	defer rows.Close()

	result := []model.Stores{}
	for rows.Next() {
		s, err := scanStore(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("[ERROR] can't scan row from stores: %w", err)
		}
		result = append(result, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("[ERROR] error query from stores: %w", err)
	}

	return result, total, nil
}

func (r *storePostgresRepository) Create(ctx context.Context, s model.Stores) (model.Stores, error) {
	var createdAt time.Time
	query := "INSERT INTO stores (tenant_id, name, description, is_active) VALUES ($1, $2, $3, $4) RETURNING id, created_at"

	if err := r.pool.QueryRow(ctx, query, s.OwnerID, s.Name, s.Description, s.IsActive).Scan(&s.ID, &createdAt); err != nil {
		if isUniqueViolation(err) {
			return model.Stores{}, ErrDuplicate
		}
		return model.Stores{}, fmt.Errorf("[ERROR] can't create store: %w", err)
	}

	s.CreatedAt = createdAt.Format(time.RFC3339)
	return s, nil
}

func (r *storePostgresRepository) Update(ctx context.Context, s model.Stores) (model.Stores, error) {
	var createdAt time.Time
	query := `UPDATE stores 
	          SET name = $1, description = $2, is_active = $3, updated_at = NOW() 
	          WHERE id = $4 
	          RETURNING id, tenant_id, name, COALESCE(description, ''), is_active, created_at`

	if err := r.pool.QueryRow(ctx, query, s.Name, s.Description, s.IsActive, s.ID).Scan(
		&s.ID, &s.OwnerID, &s.Name, &s.Description, &s.IsActive, &createdAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Stores{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Stores{}, ErrDuplicate
		}
		return model.Stores{}, fmt.Errorf("[ERROR] can't update store: %w", err)
	}

	s.CreatedAt = createdAt.Format(time.RFC3339)
	return s, nil
}

func (r *storePostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM stores WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("[ERROR] can't delete store: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func buildFilterStore(q model.ListQuery) (string, []any) {
	where := " WHERE 1=1 "
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}

	if q.StoreFilter != nil && q.StoreFilter.OwnerID > 0 {
		where += fmt.Sprintf(" AND tenant_id = $%d", len(args)+1)
		args = append(args, q.StoreFilter.OwnerID)
	}

	return where, args
}

func scanStore(rows pgx.Rows) (model.Stores, error) {
	var s model.Stores
	var createdAt time.Time

	if err := rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.Description, &s.IsActive, &createdAt); err != nil {
		return model.Stores{}, err
	}
	s.CreatedAt = createdAt.Format(time.RFC3339)

	return s, nil
}
