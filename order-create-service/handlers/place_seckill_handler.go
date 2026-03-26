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
)

func PlaceSeckill(c *gin.Context) {
	var p model.Product
	ctx := context.Background()
	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "数据格式错误",
		})
		logs.WriteLog(err)
		return
	}
	productIdStr := strconv.Itoa(p.ProductId)
	err = dao.Rdb.Set(ctx, productIdStr, p.Stock, p.EndTime.Sub(time.Now())).Err()
	if err != nil {
		logs.WriteLog(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "添加秒杀失败",
		})
		return
	}
	err = dao.Rdb.Set(ctx, productIdStr+"StartTime", p.StartTime.Unix(), p.EndTime.Sub(time.Now())).Err()
	if err != nil {
		logs.WriteLog(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "添加秒杀开始时间失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"info": "添加秒杀成功",
	})
}
