// handlers/password_reset.go
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"car-store-backend/db"

	"golang.org/x/crypto/bcrypt"
)

// একটা random, unique token তৈরি করা (৩২ বাইট → ৬৪ ক্যারেক্টার hex স্ট্রিং)
func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Gmail SMTP দিয়ে email পাঠানোর ফাংশন
func sendResetEmail(toEmail, resetLink string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	subject := "Reset Your Password"
	body := fmt.Sprintf(
		"You requested a password reset.\n\nClick the link below to set a new password (valid for 15 minutes):\n\n%s\n\nIf you didn't request this, you can safely ignore this email.",
		resetLink,
	)

	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, message)
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// POST /api/forgot-password — email নিয়ে reset token তৈরি করে email পাঠানো
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// ইউজার আছে কিনা খোঁজা
	var userID int
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		// নিরাপত্তার জন্য — email না থাকলেও একই সফল মেসেজ দেখাই,
		// যাতে কেউ এটা দিয়ে বুঝতে না পারে কোন email সিস্টেমে registered আছে
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If an account with that email exists, a reset link has been sent.",
		})
		return
	}

	token, err := generateResetToken()
	if err != nil {
		http.Error(w, "Failed to generate reset token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	_, err = db.Pool.Exec(ctx,
		"INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to create reset token", http.StatusInternalServerError)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	if err := sendResetEmail(req.Email, resetLink); err != nil {
		http.Error(w, "Failed to send reset email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If an account with that email exists, a reset link has been sent.",
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// POST /api/reset-password — token verify করে নতুন password সেট করা
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Token == "" || len(req.NewPassword) < 6 {
		http.Error(w, "Valid token and a password of at least 6 characters are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	var userID int
	var expiresAt time.Time
	err := db.Pool.QueryRow(ctx,
		"SELECT user_id, expires_at FROM password_reset_tokens WHERE token = $1",
		req.Token,
	).Scan(&userID, &expiresAt)

	if err != nil {
		http.Error(w, "Invalid or expired reset link", http.StatusBadRequest)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "This reset link has expired. Please request a new one.", http.StatusBadRequest)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process new password", http.StatusInternalServerError)
		return
	}

	_, err = db.Pool.Exec(ctx, "UPDATE users SET password_hash = $1 WHERE id = $2", string(newHash), userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// ব্যবহৃত token মুছে ফেলা, যাতে একই লিংক দ্বিতীয়বার ব্যবহার না করা যায়
	_, _ = db.Pool.Exec(ctx, "DELETE FROM password_reset_tokens WHERE token = $1", req.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password has been reset successfully"})
}