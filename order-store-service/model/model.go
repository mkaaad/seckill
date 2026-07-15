package model

// Order persisted in MySQL. Unique index prevents duplicate orders on Kafka redelivery.
type Order struct {
	ProductId int `json:"product_id" gorm:"not null;uniqueIndex:uk_user_product"`
	UserId    int `json:"user_id" gorm:"not null;uniqueIndex:uk_user_product"`
}

const (
	StatusPending = "pending"
	StatusFailed  = "failed"
)
