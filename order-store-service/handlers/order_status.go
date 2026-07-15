package handlers

import (
	"fmt"
	"time"
)

const orderStatusTTL = 15 * time.Minute

func orderStatusKey(userId, productId int) string {
	return fmt.Sprintf("order:status:%d:%d", userId, productId)
}

func stockKey(productId int) string {
	return fmt.Sprintf("%d", productId)
}
