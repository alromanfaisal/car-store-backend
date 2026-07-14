// handlers/cart_router.go
package handlers

import (
	"net/http"
	"strings"
)

// CartItemRouter — /api/cart/{id} পাথ থেকে {id} বের করে PUT/DELETE handler এ পাঠায়
func CartItemRouter(w http.ResponseWriter, r *http.Request) {
	// URL থেকে "/api/cart/" অংশটুকু বাদ দিয়ে বাকি অংশ (আইডি) বের করা
	itemID := strings.TrimPrefix(r.URL.Path, "/api/cart/")
	if itemID == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		UpdateCartItemHandler(w, r, itemID)
	case http.MethodDelete:
		DeleteCartItemHandler(w, r, itemID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}