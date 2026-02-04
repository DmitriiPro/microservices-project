package repository

import (
	"context"
	"fmt"

	"github.com/DmitriiPro/order-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userId int64, product string, quantity int64) (*model.Order, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateOrder(ctx context.Context, userId int64, product string, quantity int64) (*model.Order, error) {
	query := `INSERT INTO orders (user_id, product, quantity) VALUES ($1, $2, $3) RETURNING id, user_id, product, quantity, created_at`

	var o model.Order
	err := r.db.QueryRow(ctx, query, userId, product, quantity).Scan(&o.Id,
		&o.UserId,
		&o.Product,
		&o.Quantity,
		&o.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &o, nil
}
