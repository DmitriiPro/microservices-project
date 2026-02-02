package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) *sql.DB {
	log.Printf("Connecting to PostgreSQL with DSN: %s", dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to postgres:", err)
	}

	// Установите настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Пинг с ретраями
	var pingErr error
	for i := 0; i < 5; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		log.Printf("PostgreSQL ping attempt %d failed: %v", i+1, pingErr)
		time.Sleep(2 * time.Second)
	}

	if pingErr != nil {
		log.Fatalf("postgres ping not responding after retries: %v", pingErr)
	}

	log.Println("PostgreSQL connected successfully")
	return db
}
