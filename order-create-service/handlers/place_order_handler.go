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
	n, err := dao.Rdb.Exists(ctx, "history"+userId).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	//如果此用户在特定时间段内没有请求过，放行，并写入访问历史
	if n == 0 {
		//访问历史在特定时间后过期
		err = dao.Rdb.Set(ctx, "history"+userId, 1, time.Second).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"info": "服务器内部错误",
			})
			logs.WriteLog(err)
			return
		}
	} else {
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
	//检查库存是否有剩余，利用自减实现原子化操作
	stock, err := dao.Rdb.Decr(ctx, productId).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	//如果库存不足，把多减的库存加回来并提示库存不足
	if stock < 0 {
		dao.Rdb.Incr(ctx, productId)
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "库存不足",
		})
		return
	}
	//所有被放行的流量将写入kafka
	err = sendMessage(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "服务器内部错误",
		})
		logs.WriteLog(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"info":  "订单创建成功",
		"order": order,
	})
}
