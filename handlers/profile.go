// handlers/profile.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"car-store-backend/db"
	"car-store-backend/middleware"

	"golang.org/x/crypto/bcrypt"
)

type UserProfile struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GET /api/me — লগইন করা ইউজারের বর্তমান তথ্য
func GetMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var profile UserProfile
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1",
		userID,
	).Scan(&profile.ID, &profile.Name, &profile.Email)

	if err != nil {
		http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

type UpdateNameRequest struct {
	Name string `json:"name"`
}

// PUT /api/me — শুধু নাম আপডেট করা
func UpdateMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Pool.Exec(
		context.Background(),
		"UPDATE users SET name = $1 WHERE id = $2",
		req.Name, userID,
	)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// PUT /api/me/password — password পরিবর্তন করা
func UpdateMyPasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 6 {
		http.Error(w, "New password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// প্রথমে বর্তমান password ঠিক আছে কিনা যাচাই করা (নিরাপত্তার জন্য জরুরি —
	// নাহলে কেউ লগইন করা সেশন hijack করে সরাসরি password বদলে দিতে পারতো)
	var currentHash string
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT password_hash FROM users WHERE id = $1",
		userID,
	).Scan(&currentHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process new password", http.StatusInternalServerError)
		return
	}

	_, err = db.Pool.Exec(
		context.Background(),
		"UPDATE users SET password_hash = $1 WHERE id = $2",
		string(newHash), userID,
	)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
}