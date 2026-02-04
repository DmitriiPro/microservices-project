package client

import "time"

type User struct {
	Id        int64     `json:"id,string"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
