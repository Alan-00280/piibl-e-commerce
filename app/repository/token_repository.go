package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO
// 1. Buat Interface bernama: TokenRepository
// interface berisi:
//
//	Save(ctx context.Context, t model.RefreshToken) error
//	FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error)
//	Revoke(ctx context.Context, tokenHash string) error
//	RevokeAllForUser(ctx context.Context, userID int) error
type TokenRepository interface {
	Save(ctx context.Context, t model.RefreshToken) error
	FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID int) error
}

// 2. struct tokenPGRepository mempunyai atribut *pgxpool.Pool
type tokenPGRepository struct {
	pool *pgxpool.Pool
}

// 3. function menerima *pgxpool.Pool dan mengembalikan interface TokenRepository
func NewAuthRepo(pool *pgxpool.Pool) TokenRepository {
	return &tokenPGRepository{pool: pool}
}

// 4. Implementasi functions ke tokenPGRepository

// -- IMPLEMENTATION --
// Save() - INSERT INTO refresh_tokens
// FindActive() - WHERE token_hash = $1
// Revoke() - UPDATE refresh_tokens SET revoked_at = $1 WHERE token_hash = $2
// RevokeAllForUser() - Sama seperti revoke tapi WHERE user_id = $1

func (r *tokenPGRepository) Save(ctx context.Context, t model.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, t.UserID, t.TokenHash, t.ExpiredAt)

	if err != nil {
		return fmt.Errorf("[ERROR] can't save refresh token: %w", err)
	}

	return nil
}

func (r *tokenPGRepository) FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	var t model.RefreshToken

	if err := r.pool.QueryRow(ctx, `SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`, tokenHash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiredAt, &t.RevokedAt, &t.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}

		return model.RefreshToken{}, fmt.Errorf("can't get token: %w", err)
	}

	return t, nil
}

func (r *tokenPGRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash)

	if err != nil {
		return fmt.Errorf("[ERROR] can't revoke token: %w", err)
	}

	return nil
}

func (r *tokenPGRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)

	if err != nil {
		return fmt.Errorf("[ERROR] can't revoke token for user: %w", err)
	}

	return nil
}
