package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// interface RoleRepository
// method:
//
//	LoadPermissions(ctx context.Context) (map[string][]string)
type RoleRepository interface {
	LoadPermissions(ctx context.Context) (map[string][]string, error)
}

// struct yang akan implement interface RoleRepository
// rolePostgresRepository attribut:
//
//	pool *pgxpool.Pool
type rolePostgresRepository struct {
	pool *pgxpool.Pool
}

// constructor
func NewRoleRepository(pool *pgxpool.Pool) RoleRepository {
	return &rolePostgresRepository{pool: pool}
}

// methods implementations
// LoadPermissions()
//
//	mengambil semua roles (baik ada permission atau tidak)
//	kembalian berupa map[string][]string
//	  ex: {'admin' = {'user:list', 'user:update:any'}, 'staff'={'user:list'}}
func (r *rolePostgresRepository) LoadPermissions(ctx context.Context) (map[string][]string, error) {
	result := make(map[string][]string)

	sql_string := `
		SELECT r.name, COALESCE(permission_name,'')
		FROM roles r
		LEFT JOIN role_permissions rp ON r.name = rp.role_name
		ORDER BY r.name, rp.permission_name
	`

	rows, err := r.pool.Query(ctx, sql_string)
	if err != nil {
		return nil, fmt.Errorf("can't get permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role, permission string

		if err := rows.Scan(&role, &permission); err != nil {
			return nil, fmt.Errorf("can't read row permission: %w", err)
		}

		_, ok := result[role]
		if !ok {
			result[role] = []string{}
		}

		if permission != "" {
			result[role] = append(result[role], permission)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("can't read role permissions: %w", err)
	}

	return result, nil
}
