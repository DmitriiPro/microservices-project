package model

import "time"

type Order struct {
	Id        int64     `db:"id"`
	UserId    int64     `db:"user_id"`
	Product   string    `db:"product"`
	Quantity  int64     `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
}
