package middleware

import (
	"backend/config"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_key")

type contextKey string

const UserIDKey contextKey = "user_id"

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CheckPermission(userID int, permission string) (bool, error) {
	// Check user for admin permission
	var userType string
	err := config.PostgresDB.QueryRow("SELECT type FROM Users WHERE user_id = $1", userID).Scan(&userType)
	if err != nil {
		return false, err
	}
	if userType == "admin" {
		return true, nil
	}

	// Check user for permission
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM Role_Permissions rp
			JOIN Users u ON u.role_id = rp.role_id
			JOIN Permissions p ON p.permission_id = rp.permission_id
			WHERE u.user_id = $1 AND p.name = $2
		)
	`
	err = config.PostgresDB.QueryRow(query, userID, permission).Scan(&exists)
	return exists, err
}

func RequirePermission(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		hasPermission, err := CheckPermission(userID, permission)
		if err != nil || !hasPermission {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
