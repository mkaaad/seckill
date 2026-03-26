package main

import (
	"order-store/dao"
	"order-store/handlers"
	"order-store/logs"
)

func main() {
	dao.ClientDB()
	logs.OpenFile()
	handlers.WriteToMysql()
}
