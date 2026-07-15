package handlers

import (
	"fmt"
	"time"
)

const orderStatusTTL = 15 * time.Minute
const orderSuccessCacheTTL = 5 * time.Minute

// orderStatusKey builds the Redis key for order status cache.
// Format: order:status:{userId}:{productId}
func orderStatusKey(userId, productId int) string {
	return fmt.Sprintf("order:status:%d:%d", userId, productId)
}

// stockKey matches place_order / place_seckill inventory key (product id string).
func stockKey(productId int) string {
	return fmt.Sprintf("%d", productId)
}
