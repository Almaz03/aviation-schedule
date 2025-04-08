package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func Init() *sql.DB {
	dsn := "postgres://postgres:postgres@localhost:5432/flights?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	return db
}
