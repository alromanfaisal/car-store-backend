// middleware/admin.go
package middleware

import (
	"context"
	"net/http"

	"car-store-backend/db"
)

// RequireAdmin — এটা RequireAuth এর পরে ব্যবহার হবে (chain করে), যাতে
// প্রথমে token verify হয়, তারপর সেই ইউজার আসলেই admin কিনা চেক হয়
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var isAdmin bool
		err := db.Pool.QueryRow(context.Background(),
			"SELECT is_admin FROM users WHERE id = $1", userID,
		).Scan(&isAdmin)

		if err != nil || !isAdmin {
			http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}