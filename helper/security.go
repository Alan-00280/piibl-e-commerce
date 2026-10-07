package helper

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	bcryptCost        = 12
)

var dummyHash = []byte("$2a$12$abcdede.4fYV4cZrTHuoDB.MvJzRn..01KYucHpKikuOEcCKKK")

type PasswordCommonSet struct {
	PasswordSet map[string]struct{}
}

func NewPasswordCommonSet(path string) (*PasswordCommonSet, error) {
	passwordSet := make(map[string]struct{})

	password_path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(password_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		weak_password := scanner.Text()
		passwordSet[weak_password] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &PasswordCommonSet{
		PasswordSet: passwordSet,
	}, nil
}

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)

	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func VerifyDummyPassword(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}

func RandomToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func SHA256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
