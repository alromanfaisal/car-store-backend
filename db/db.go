// db/db.go
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

	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable" // লোকাল ডেভেলপমেন্টে ডিফল্ট
	}

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode,
	)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	Pool = pool
	log.Println("Database connected successfully!")
}