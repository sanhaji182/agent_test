// Package auth menyediakan autentikasi JWT dan manajemen API key
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrUnauthorized dikembalikan saat token tidak valid
var ErrUnauthorized = errors.New("unauthorized")

// Claims adalah payload JWT yang berisi info user
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Auth mengelola JWT token generation dan validation
type Auth struct {
	jwtSecret []byte
}

// New membuat Auth baru dengan secret key
func New(secret string) *Auth {
	return &Auth{jwtSecret: []byte(secret)}
}

// GenerateToken membuat JWT token baru untuk user (berlaku 24 jam)
func (a *Auth) GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateToken memvalidasi JWT token dan mengembalikan claims
func (a *Auth) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrUnauthorized
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

// Middleware adalah HTTP middleware yang memvalidasi Bearer token
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := a.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type claimsKey struct{}

// GetClaims mengambil claims dari context (setelah melewati middleware)
func GetClaims(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey{}).(*Claims)
	return c
}

// --- Manajemen API Key ---

// GenerateAPIKey membuat API key baru (plain + hash untuk disimpan di DB)
func GenerateAPIKey() (plain string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	plain = "gta_" + hex.EncodeToString(b) // Prefix "gta_" untuk identifikasi
	hash = HashAPIKey(plain)
	return plain, hash, nil
}

// HashAPIKey menghasilkan SHA-256 hash dari API key
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
