package auth


import (
//	"fmt"
	"net/http"
	"github.com/golang-jwt/jwt/v5"
	
)
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenStr := r.Header.Get("Authorization")
        if tokenStr == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        // Parse and validate JWT
        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            return []byte("secret"), nil // In real life, use a secure key
        })
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}