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
	if err := godotenv.Load(); err != nil {
	log.Println("No .env file found, using system environment variables instead")
}

	db.Connect()
	defer db.Pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/signup", handlers.SignupHandler)
	mux.HandleFunc("/api/login", handlers.LoginHandler)

	mux.HandleFunc("/api/me", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetMyProfileHandler(w, r)
		case http.MethodPut:
			handlers.UpdateMyProfileHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/me/password", middleware.RequireAuth(handlers.UpdateMyPasswordHandler))

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

	mux.HandleFunc("/api/orders", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetOrdersHandler(w, r)
		case http.MethodPost:
			handlers.CreateOrderHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	handler := middleware.EnableCORS(mux)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}