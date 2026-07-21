// middleware/cors.go
package middleware

import (
	"net/http"
	"os"
)

// EnableCORS — Frontend থেকে আসা request গুলো allow করার জন্য
// FRONTEND_URL environment variable থেকে origin পড়ে, না থাকলে লোকাল ডেভেলপমেন্টের ডিফল্ট ব্যবহার করে
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("FRONTEND_URL")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Browser প্রথমে একটা "preflight" OPTIONS request পাঠায়, সেটা এখানেই শেষ করে দিচ্ছি
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}