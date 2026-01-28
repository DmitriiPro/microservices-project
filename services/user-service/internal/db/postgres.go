package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to postgres:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("postgres ping not responding: %v", err)
	}

	fmt.Println("Postgres connected")
	return db
}
