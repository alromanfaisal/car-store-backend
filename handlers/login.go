// handlers/login.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"car-store-backend/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// ডাটাবেজ থেকে এই email এর ইউজার খোঁজা
	var userID int
	var name string
	var passwordHash string
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, name, password_hash FROM users WHERE email = $1",
		req.Email,
	).Scan(&userID, &name, &passwordHash)

	if err != nil {
		// ইচ্ছাকৃতভাবে জেনেরিক error দিচ্ছি — এটা যেন কেউ বুঝতে না পারে
		// "email নেই" নাকি "password ভুল" — এটা security best practice
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// পাঠানো password, ডাটাবেজে সেভ করা hash এর সাথে মেলানো
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// password ঠিক থাকলে — JWT টোকেন তৈরি করা
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   req.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // ২৪ ঘণ্টা পর টোকেন মেয়াদোত্তীর্ণ
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// রেসপন্স পাঠানো
	var resp LoginResponse
	resp.Token = signedToken
	resp.User.ID = userID
	resp.User.Name = name
	resp.User.Email = req.Email

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}