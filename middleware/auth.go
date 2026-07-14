// middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// context এ user_id রাখার জন্য একটা কাস্টম key type
// (সাধারণ string key ব্যবহার না করার কারণ — অন্য প্যাকেজের সাথে নাম-সংঘর্ষ এড়ানো)
type contextKey string

const UserIDKey contextKey = "user_id"

// RequireAuth — এটা একটা "middleware wrapper"
// যেকোনো handler কে এর ভেতরে মুড়ে দিলে, সেই handler এর আগে JWT ভেরিফাই হবে
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// "Authorization: Bearer <token>" হেডার পড়া
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// "Bearer " অংশটুকু বাদ দিয়ে শুধু টোকেন বের করা
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// টোকেন পার্স ও verify করা
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// টোকেনের ভেতর থেকে user_id বের করা
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64) // JWT numbers সবসময় float64 হিসেবে আসে
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		userID := int(userIDFloat)

		// user_id কে request এর context এ বসিয়ে দেওয়া, যাতে পরের handler এটা পড়তে পারে
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// আসল handler কে চালানো, নতুন context সহ
		next(w, r.WithContext(ctx))
	}
}