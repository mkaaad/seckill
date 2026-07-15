package dao

import (
	"context"
	"log"
	"order-store/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB
var Rdb *redis.Client

func ClientDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/SecKill?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(&model.Order{})
	Db = db
	log.Println("MySQL 连接成功")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalln("Redis 连接失败，", err)
	}
	Rdb = rdb
	log.Println("Redis 连接成功")
}
