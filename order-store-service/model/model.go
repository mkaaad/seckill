package model

type Order struct {
	ProductId int `json:"product_id" gorm:"not null"`
	UserId    int `json:"user_id" gorm:"not null"`
}
