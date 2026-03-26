package dao

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

func ClientDB() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("数据库链接失败，", err)
	} else {
		log.Println("数据库链接成功。")
	}
	Rdb = rdb
}
