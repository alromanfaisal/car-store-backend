//car-store-backend/db/db.go
package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool — পুরো অ্যাপ জুড়ে এই একটাই কানেকশন পুল ব্যবহার হবে
var Pool *pgxpool.Pool

func Connect() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	Pool = pool
	log.Println("Database connected successfully!")
}