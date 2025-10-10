package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(dbURL string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB not reachable: %v", err)
	}

	log.Println("Connected to PostgreSQL ✅")
	return db
}
