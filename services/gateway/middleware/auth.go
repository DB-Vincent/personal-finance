package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DB-Vincent/personal-finance/pkg/response"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			tokenString, ok := strings.CutPrefix(header, "Bearer ")
			if !ok {
				response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return secret, nil
			})
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(*tokenClaims)
			if !ok || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			r.Header.Set("X-User-ID", claims.UserID.String())
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)

			next.ServeHTTP(w, r)
		})
	}
}
