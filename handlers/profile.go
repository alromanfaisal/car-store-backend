// handlers/profile.go
package handlers

import (
	"encoding/json"
	"net/http"

	"car-store-backend/middleware"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// middleware যে user_id context এ বসিয়ে দিয়েছিল, সেটা এখানে পড়া হচ্ছে
	userID := r.Context().Value(middleware.UserIDKey).(int)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "This is a protected route!",
		"user_id": userID,
	})
}