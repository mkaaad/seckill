package handlers

import (
	"context"
	"net/http"
	"order-create/dao"
	"order-create/logs"
	"order-create/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// time import still used by startTime check below

// Atomic stock deduct: DECR + restore if negative in one Lua script.
const deductStockLua = `
local s = redis.call('DECR', KEYS[1])
if s < 0 then
  redis.call('INCR', KEYS[1])
  return -1
end
return s
`

func PlaceOrder(c *gin.Context) {
	ctx := context.Background()
	var order model.Order
	productId := c.Query("product_id")
	userId := c.Query("user_id")
	if productId == "" || userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "用户id或商品id不能为空",
		})
		return
	}
	// 令牌桶限流（每用户 capacity=5 突发，rate=1 token/s 补充）
	ok, _, err := allowTokenBucket(ctx, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	if !ok {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"info": "请求过于频繁，稍后再试",
		})
		return
	}
	startTimeStr, err := dao.Rdb.Get(ctx, productId+"StartTime").Result()
	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "该商品未在秒杀",
		})
		return

	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	if time.Now().Unix() < startTime {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "秒杀未开始",
		})
		return
	}
	//判断请求中的参数是否可以转化为整数
	order.ProductId, err = strconv.Atoi(productId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "参数不符合要求",
		})
		return
	}
	order.UserId, err = strconv.Atoi(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "参数不符合要求",
		})
		return
	}

	// Lua 原子扣库存（避免 DECR 后 INCR 回补的竞态）
	stock, err := dao.Rdb.Eval(ctx, deductStockLua, []string{productId}).Int64()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	if stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "库存不足",
		})
		return
	}

	// 投递 Kafka；失败则回补库存，不写 pending
	err = sendMessage(order)
	if err != nil {
		if _, incrErr := dao.Rdb.Incr(ctx, productId).Result(); incrErr != nil {
			logs.WriteLog(incrErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	// Kafka 成功后写入 pending 状态缓存；落库成功后由 store 删缓存，失败则改为 failed
	if err = dao.Rdb.Set(ctx, orderStatusKey(order.UserId, order.ProductId), model.StatusPending, orderStatusTTL).Err(); err != nil {
		logs.WriteLog(err)
		// 缓存失败不阻断下单响应，查询可回源 MySQL
	}
	c.JSON(http.StatusOK, gin.H{
		"info":   "订单创建成功",
		"status": model.StatusPending,
		"order":  order,
	})
}
