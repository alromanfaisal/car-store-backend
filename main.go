// main.go
package main

import (
	"log"
	"net/http"

	"car-store-backend/db"
	"car-store-backend/handlers"
	"car-store-backend/middleware"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	db.Connect()
	defer db.Pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/signup", handlers.SignupHandler)
	mux.HandleFunc("/api/login", handlers.LoginHandler)
	mux.HandleFunc("/api/profile", middleware.RequireAuth(handlers.ProfileHandler))

	// Products — এগুলো public, লগইন ছাড়াই দেখা যাবে
	mux.HandleFunc("/api/products", handlers.GetProductsHandler)
	mux.HandleFunc("/api/products/", handlers.GetProductByIDHandler)

	mux.HandleFunc("/api/cart", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetCartHandler(w, r)
		case http.MethodPost:
			handlers.AddToCartHandler(w, r)
		case http.MethodDelete:
			handlers.ClearCartHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/cart/", middleware.RequireAuth(handlers.CartItemRouter))

	handler := middleware.EnableCORS(mux)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}