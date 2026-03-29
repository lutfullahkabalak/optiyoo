package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"optiyoo-backend/config"
)

type ctxKey int

const userIDKey ctxKey = 1

// UserIDFromContext returns the authenticated user id set by RequireAuth.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

var ErrNoAuth = errors.New("no authorization bearer")

// ParseBearerUserID validates the JWT from Authorization: Bearer and returns the subject (user id).
func ParseBearerUserID(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", ErrNoAuth
	}
	raw := strings.TrimSpace(strings.TrimPrefix(h, prefix))
	if raw == "" {
		return "", ErrNoAuth
	}
	secret := config.JWTSecret()
	tok, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !tok.Valid {
		return "", ErrNoAuth
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrNoAuth
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", ErrNoAuth
	}
	return sub, nil
}

// SignUserToken issues a short-lived HS256 JWT for the given user id.
func SignUserToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(7 * 24 * time.Hour).Unix(),
		"iss": "optiyoo",
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(config.JWTSecret()))
}

// RequireAuth rejects requests without a valid Bearer JWT and stores sub in context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, err := ParseBearerUserID(r)
		if err != nil || uid == "" {
			http.Error(w, "Oturum gerekli.", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
