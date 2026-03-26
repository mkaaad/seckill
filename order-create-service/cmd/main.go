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
	r.GET("/order", handlers.PlaceOrder)
	r.Run(":8080")
}
