package handlers

import (
	"context"
	"errors"
	"net/http"
	"order-create/dao"
	"order-create/logs"
	"order-create/model"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// SearchOrder returns order status: Redis first, MySQL on cache miss.
// GET /order/search?user_id=&product_id=
func SearchOrder(c *gin.Context) {
	ctx := context.Background()
	userIdStr := c.Query("user_id")
	productIdStr := c.Query("product_id")
	if userIdStr == "" || productIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"info":   "用户id或商品id不能为空",
			"status": model.StatusNotFound,
		})
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info":   "参数不符合要求",
			"status": model.StatusNotFound,
		})
		return
	}
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info":   "参数不符合要求",
			"status": model.StatusNotFound,
		})
		return
	}

	order := model.Order{UserId: userId, ProductId: productId}
	key := orderStatusKey(userId, productId)

	// 1) Redis hit → pending / failed (or any cached value)
	cached, err := dao.Rdb.Get(ctx, key).Result()
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"info":   "ok",
			"status": cached,
			"order":  order,
		})
		return
	}
	if err != redis.Nil {
		logs.WriteLog(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		return
	}

	// 2) Cache miss → MySQL
	var row model.Order
	result := dao.Db.Where("user_id = ? AND product_id = ?", userId, productId).First(&row)
	if result.Error == nil {
		// optional short cache of success (not the primary update path)
		_ = dao.Rdb.Set(ctx, key, model.StatusSuccess, orderSuccessCacheTTL).Err()
		c.JSON(http.StatusOK, gin.H{
			"info":   "ok",
			"status": model.StatusSuccess,
			"order":  order,
		})
		return
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"info":   "订单不存在",
			"status": model.StatusNotFound,
		})
		return
	}
	logs.WriteLog(result.Error)
	c.JSON(http.StatusInternalServerError, gin.H{
		"info": "服务器内部错误",
	})
}
