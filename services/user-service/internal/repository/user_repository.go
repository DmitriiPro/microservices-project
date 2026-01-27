package repository

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID        int64
	Email     string
	CreatedAt time.Time
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email string) (int64, error) {
	query := `INSERT INTO users (email) VALUES ($1) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, email).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, email, created_at FROM users WHERE id = $1`	
	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}