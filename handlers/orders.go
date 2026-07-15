
// handlers/orders.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"car-store-backend/db"
	"car-store-backend/middleware"
)

type OrderItem struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
}

type Order struct {
	ID           int         `json:"id"`
	CustomerName string      `json:"customer_name"`
	Phone        string      `json:"phone"`
	Address      string      `json:"address"`
	TotalPrice   int         `json:"total_price"`
	Status       string      `json:"status"`
	CreatedAt    string      `json:"created_at"`
	Items        []OrderItem `json:"items"`
}

type CreateOrderRequest struct {
	CustomerName string `json:"customer_name"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
}

// POST /api/orders — cart এর ভিত্তিতে নতুন অর্ডার তৈরি করা
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.CustomerName == "" || req.Phone == "" || req.Address == "" {
		http.Error(w, "Name, phone, and address are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// ধাপ ১: cart_items আর cars জয়েন করে cart এর বর্তমান আইটেমগুলো + দাম বের করা
	rows, err := db.Pool.Query(ctx, `
		SELECT c.product_id, cars.name, c.quantity,
		       COALESCE(cars.discount_price, cars.price) AS unit_price
		FROM cart_items c
		JOIN cars ON cars.id = c.product_id
		WHERE c.user_id = $1
	`, userID)
	if err != nil {
		http.Error(w, "Failed to read cart", http.StatusInternalServerError)
		return
	}

	type cartRow struct {
		ProductID int
		Name      string
		Quantity  int
		UnitPrice int
	}
	var cartRows []cartRow
	totalPrice := 0

	for rows.Next() {
		var cr cartRow
		if err := rows.Scan(&cr.ProductID, &cr.Name, &cr.Quantity, &cr.UnitPrice); err != nil {
			rows.Close()
			http.Error(w, "Failed to read cart item", http.StatusInternalServerError)
			return
		}
		cartRows = append(cartRows, cr)
		totalPrice += cr.UnitPrice * cr.Quantity
	}
	rows.Close()

	if len(cartRows) == 0 {
		http.Error(w, "Your cart is empty", http.StatusBadRequest)
		return
	}

	// ধাপ ২: একটা transaction শুরু করা — যাতে order আর সব order_items একসাথে
	// সফল হয়, নাহলে কিছুই সেভ না হয় (partial/broken অর্ডার এড়ানোর জন্য)
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx) // কোথাও error হলে সব বাতিল হয়ে যাবে

	var orderID int
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, customer_name, phone, address, total_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, req.CustomerName, req.Phone, req.Address, totalPrice).Scan(&orderID)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	for _, cr := range cartRows {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, quantity, price)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, cr.ProductID, cr.Name, cr.Quantity, cr.UnitPrice)
		if err != nil {
			http.Error(w, "Failed to save order items", http.StatusInternalServerError)
			return
		}
	}

	// ধাপ ৩: cart খালি করে দেওয়া (অর্ডার সফল হলে)
	_, err = tx.Exec(ctx, "DELETE FROM cart_items WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to finalize order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Order placed successfully",
		"order_id": orderID,
	})
}

// GET /api/orders — লগইন করা ইউজারের সব অর্ডার (নতুন থেকে পুরোনো)
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	ctx := context.Background()

	orderRows, err := db.Pool.Query(ctx, `
    SELECT id, customer_name, phone, address, total_price, status, created_at::text
    FROM orders
    WHERE user_id = $1
    ORDER BY created_at DESC
`, userID)
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	defer orderRows.Close()

	orders := []Order{}
	for orderRows.Next() {
		var o Order
		if err := orderRows.Scan(&o.ID, &o.CustomerName, &o.Phone, &o.Address, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			http.Error(w, "Failed to read order", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	// প্রতিটা অর্ডারের items আলাদাভাবে নিয়ে আসা
	for i := range orders {
		itemRows, err := db.Pool.Query(ctx, `
			SELECT product_id, product_name, quantity, price
			FROM order_items
			WHERE order_id = $1
		`, orders[i].ID)
		if err != nil {
			http.Error(w, "Failed to fetch order items", http.StatusInternalServerError)
			return
		}

		items := []OrderItem{}
		for itemRows.Next() {
			var item OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &item.Price); err != nil {
				itemRows.Close()
				http.Error(w, "Failed to read order item", http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		itemRows.Close()
		orders[i].Items = items
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

var _ = strings.TrimSpace // placeholder import guard