package helper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token sudah kedaluarsa")
)

type accessClaims struct {
	Username string     `json:"username"`
	Role     model.Role `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewJWTManager(secret string, issuer string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		issuer:    issuer,
		accessTTL: accessTTL,
	}
}

func (m *JWTManager) GetTTL() time.Duration {
	return m.accessTTL
}

func (m *JWTManager) GenerateAccessToken(u model.User) (string, error) {
	now := time.Now()

	claims := accessClaims{
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(u.ID),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("can't create token: %w", err)
	}

	return signed, nil
}

func (m *JWTManager) Parse(tokenString string) (model.AuthUser, error) {
	claims := &accessClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algorithm isn't valid: %v", t.Header["alg"])
			}
			return m.secret, nil
		}, jwt.WithIssuer(m.issuer), jwt.WithExpirationRequired())

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return model.AuthUser{}, ErrExpiredToken
		}

		return model.AuthUser{}, ErrInvalidToken
	}

	if !token.Valid {
		return model.AuthUser{}, ErrInvalidToken
	}

	userID, err := strconv.Atoi(claims.Subject)

	return model.AuthUser{
		UserID:   userID,
		Role:     claims.Role,
		Username: claims.Username,
	}, nil
}
