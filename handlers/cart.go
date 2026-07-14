// handlers/cart.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"car-store-backend/db"
	"car-store-backend/middleware"

	"github.com/jackc/pgx/v5"
)

type CartItem struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type AddToCartRequest struct {
	ProductID int `json:"product_id"`
}

// GET /api/cart — লগইন করা ইউজারের পুরো cart ফেরত দেওয়া
func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	rows, err := db.Pool.Query(
		context.Background(),
		"SELECT id, product_id, quantity FROM cart_items WHERE user_id = $1 ORDER BY created_at",
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch cart", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []CartItem{}
	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity); err != nil {
			http.Error(w, "Failed to read cart data", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// POST /api/cart — নতুন প্রোডাক্ট যোগ করা (আগে থেকে থাকলে quantity +1 বাড়ানো)
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ProductID == 0 {
		http.Error(w, "product_id is required", http.StatusBadRequest)
		return
	}

	// UPSERT — থাকলে quantity বাড়ানো, না থাকলে নতুন row বানানো
	// এটাই সেই জায়গা যেখানে আগের UNIQUE(user_id, product_id) constraint কাজে লাগছে
	_, err := db.Pool.Exec(
		context.Background(),
		`INSERT INTO cart_items (user_id, product_id, quantity)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (user_id, product_id)
		 DO UPDATE SET quantity = cart_items.quantity + 1`,
		userID, req.ProductID,
	)
	if err != nil {
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item added to cart"})
}

// PUT /api/cart/{id} — quantity সরাসরি আপডেট করা (increase/decrease বাটনের জন্য)
func UpdateCartItemHandler(w http.ResponseWriter, r *http.Request, itemID string) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		// quantity ০ বা তার কম হলে item টাই ডিলিট করে দেওয়া
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM cart_items WHERE id = $1 AND user_id = $2",
			itemID, userID,
		)
		if err != nil {
			http.Error(w, "Failed to remove item", http.StatusInternalServerError)
			return
		}
	} else {
		_, err := db.Pool.Exec(
			context.Background(),
			"UPDATE cart_items SET quantity = $1 WHERE id = $2 AND user_id = $3",
			req.Quantity, itemID, userID,
		)
		if err != nil {
			http.Error(w, "Failed to update item", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Cart updated"})
}

// DELETE /api/cart/{id} — নির্দিষ্ট item সম্পূর্ণ মুছে ফেলা
func DeleteCartItemHandler(w http.ResponseWriter, r *http.Request, itemID string) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	tag, err := db.Pool.Exec(
		context.Background(),
		"DELETE FROM cart_items WHERE id = $1 AND user_id = $2",
		itemID, userID,
	)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item removed"})
}

// DELETE /api/cart — পুরো cart খালি করা (order placement এর পর ব্যবহার হবে)
func ClearCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	_, err := db.Pool.Exec(
		context.Background(),
		"DELETE FROM cart_items WHERE user_id = $1",
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Cart cleared"})
}

var _ = pgx.ErrNoRows // placeholder import guard, ignored