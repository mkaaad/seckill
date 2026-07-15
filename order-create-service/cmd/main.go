package main

import (
	"order-create/dao"
	"order-create/handlers"
	"order-create/logs"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	dao.ClientDB()
	logs.OpenFile()
	r.POST("/seckill", handlers.PlaceSeckill)
	// /order/search 必须在 /order 之前注册，避免被更宽路径干扰
	r.GET("/order/search", handlers.SearchOrder)
	r.GET("/order", handlers.PlaceOrder)
	r.Run(":8080")
}
