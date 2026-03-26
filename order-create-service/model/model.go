package model

import (
	"time"
)

type Product struct {
	ProductId int       `json:"product_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Price     float32   `json:"price"`
	Stock     int       `json:"stock"`
}
type Order struct {
	ProductId int `json:"product_id"`
	UserId    int `json:"user_id"`
}
