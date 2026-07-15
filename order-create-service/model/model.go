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

// Order is the payload sent to Kafka and stored in MySQL.
// Unique (user_id, product_id) enforces at-most-one order per user per product.
type Order struct {
	ProductId int `json:"product_id" gorm:"not null;uniqueIndex:uk_user_product"`
	UserId    int `json:"user_id" gorm:"not null;uniqueIndex:uk_user_product"`
}

// Order status values for Redis cache / search API.
const (
	StatusPending  = "pending"
	StatusSuccess  = "success"
	StatusFailed   = "failed"
	StatusNotFound = "not_found"
)
