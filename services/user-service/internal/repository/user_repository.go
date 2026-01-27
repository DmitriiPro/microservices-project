package repository

import (
	"context"
	"database/sql"

	"github.com/DmitriiPro/user-service/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateUser(ctx context.Context, email string) (int64, error) {
	query := `INSERT INTO users (email) VALUES ($1) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, email).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *postgresRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, email, created_at FROM users WHERE id = $1`	
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}