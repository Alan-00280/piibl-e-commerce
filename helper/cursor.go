package helper

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
)

var (
	ErrInvalidCursor = errors.New("cursor tidak sesuai format")
)

// EncodeCursor mengubah penanda menjadi satu string yang aman di URL.
//
// base64 dipakai agar bentuk internalnya dapat kita ubah
// tanpa mengubah kontrak API
func EncodeCursor(createdAt time.Time, id int) string {
	raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) + "|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor membaca kembali penanda dari string.
//
// Seluruh jalur kegagalan mengembalikan error, tidak ada yang "diperbaiki
// diam-diam". Cursor rusak berarti permintaan client memang salah, dan
// client berhak tahu itu lewat 400
func DecodeCursor(encoded string) (model.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.Cursor{}, ErrInvalidCursor
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return model.Cursor{}, ErrInvalidCursor
	}

	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 0 {
		return model.Cursor{}, ErrInvalidCursor
	}

	return model.Cursor{
		CreatedAt: time.Unix(0, nanos).UTC(),
		ID:        id,
	}, nil
}
