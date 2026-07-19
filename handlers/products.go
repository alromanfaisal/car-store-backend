// handlers/products.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"log"
	"car-store-backend/db"
)

type Car struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	Price         int    `json:"price"`
	DiscountPrice *int   `json:"discount_price,omitempty"` // nullable, তাই pointer
	IsNew         bool   `json:"is_new"`
}

// GET /api/products — সব গাড়ির লিস্ট
func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(
    context.Background(),
    "SELECT id, name, description, image_url, price, discount_price, is_new FROM cars ORDER BY id",
)
if err != nil {
    log.Println("GetProductsHandler query error:", err)
    http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
    return
}
	defer rows.Close()

	cars := []Car{}
	for rows.Next() {
		var c Car
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ImageURL, &c.Price, &c.DiscountPrice, &c.IsNew); err != nil {
			http.Error(w, "Failed to read product data", http.StatusInternalServerError)
			return
		}
		cars = append(cars, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cars)
}

// GET /api/products/{id} — একটা নির্দিষ্ট গাড়ির বিস্তারিত
func GetProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/products/")
	if id == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	var c Car
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, name, description, image_url, price, discount_price, is_new FROM cars WHERE id = $1",
		id,
	).Scan(&c.ID, &c.Name, &c.Description, &c.ImageURL, &c.Price, &c.DiscountPrice, &c.IsNew)

	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}