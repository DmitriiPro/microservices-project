package config

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func NewPostgresConn() (*sql.DB, error) {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=user_service sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}
