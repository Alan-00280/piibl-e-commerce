package model

import "time"

type RegisterReq struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
}

type LoginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiredIn    int    `json:"expired_in"` // in seconds
}

type RefreshToken struct {
	ID        int64
	UserID    int
	TokenHash string
	ExpiredAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}
