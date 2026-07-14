// middleware/cors.go
package middleware

import "net/http"

// EnableCORS — Next.js (localhost:3000) থেকে আসা request গুলো allow করার জন্য
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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