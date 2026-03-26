package dao

import (
	"log"
	"order-store/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func ClientDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/SecKill?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(&model.Order{})
	Db = db
}
