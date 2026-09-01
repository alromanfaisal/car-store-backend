// handlers/admin_cars.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"car-store-backend/db"
)

type CreateCarRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	Price         int    `json:"price"`
	DiscountPrice *int   `json:"discount_price"`
	IsNew         bool   `json:"is_new"`
}

// POST /api/admin/cars — নতুন গাড়ি যোগ করা
func CreateCarHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.ImageURL == "" || req.Price <= 0 {
		http.Error(w, "Name, image_url, and a valid price are required", http.StatusBadRequest)
		return
	}

	var newID int
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO cars (name, description, image_url, price, discount_price, is_new)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.Name, req.Description, req.ImageURL, req.Price, req.DiscountPrice, req.IsNew,
	).Scan(&newID)

	if err != nil {
		http.Error(w, "Failed to create car", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "message": "Car created successfully"})
}

// PUT /api/admin/cars/{id} — গাড়ির তথ্য আপডেট করা
func UpdateCarHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/cars/")
	if id == "" {
		http.Error(w, "Car ID is required", http.StatusBadRequest)
		return
	}

	var req CreateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.ImageURL == "" || req.Price <= 0 {
		http.Error(w, "Name, image_url, and a valid price are required", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(context.Background(),
		`UPDATE cars SET name = $1, description = $2, image_url = $3, price = $4, discount_price = $5, is_new = $6
		 WHERE id = $7`,
		req.Name, req.Description, req.ImageURL, req.Price, req.DiscountPrice, req.IsNew, id,
	)
	if err != nil {
		http.Error(w, "Failed to update car", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Car not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Car updated successfully"})
}

// DELETE /api/admin/cars/{id} — গাড়ি মুছে ফেলা
func DeleteCarHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/cars/")
	if id == "" {
		http.Error(w, "Car ID is required", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(context.Background(), "DELETE FROM cars WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete car", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Car not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Car deleted successfully"})
}

// AdminCarItemRouter — /api/admin/cars/{id} পাথ থেকে PUT/DELETE handler এ পাঠায়
func AdminCarItemRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		UpdateCarHandler(w, r)
	case http.MethodDelete:
		DeleteCarHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}