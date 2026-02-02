package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/DmitriiPro/user-service/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, password_hash string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresRepository{db: db}
}

var ErrNotFoundUser = errors.New("user not found")

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFoundUser // пользователя нет
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresRepository) CreateUser(ctx context.Context, email, password_hash string) (*model.User, error) {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, email, password_hash).Scan(&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt)

	if err != nil {
		log.Printf("Repository: Error creating user: %v", err)
		return nil, err
	}

	return &user, nil
}


func (r *postgresRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFoundUser
		}
		log.Printf("Repository: Error getting user by ID %d: %v", id, err)
		return nil, err
	}

	return &user, nil
}
