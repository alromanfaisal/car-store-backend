//car-store-backend/handlers/signup.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"car-store-backend/db"

	"golang.org/x/crypto/bcrypt"
)

// ফ্রন্টএন্ড থেকে যে JSON আসবে, তার গঠন
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// সফল হলে যে JSON ফেরত পাঠানো হবে (password_hash বাদ দিয়ে, নিরাপত্তার জন্য)
type SignupResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// শুধু POST request গ্রহণ করা হবে
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// সাধারণ validation
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Password hash করা — এখানে raw password কখনো ডাটাবেজে যাচ্ছে না
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// ডাটাবেজে নতুন ইউজার insert করা
	var newUser SignupResponse
	err = db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, name, email",
		req.Name, req.Email, string(hashedPassword),
	).Scan(&newUser.ID, &newUser.Name, &newUser.Email)

	if err != nil {
		// যদি email আগে থেকেই থাকে (UNIQUE constraint ভায়োলেশন)
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// সফল রেসপন্স পাঠানো
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}