package dao

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Rdb *redis.Client
var Db *gorm.DB

func ClientDB() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("Redis 连接失败，", err)
	}
	log.Println("Redis 连接成功。")
	Rdb = rdb

	// 只读查单，不 AutoMigrate（表由 order-store 维护）
	dsn := "root:123456@tcp(127.0.0.1:3306)/SecKill?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln("MySQL 连接失败，", err)
	}
	log.Println("MySQL 连接成功。")
	Db = db
}
