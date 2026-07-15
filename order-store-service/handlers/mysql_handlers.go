package handlers

import (
	"context"
	"encoding/json"
	"log"
	"order-store/dao"
	"order-store/logs"
	"order-store/model"
)

func WriteToMysql() {
	var order model.Order
	partitionConsumer, err := ReadMessage()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("连接kafka服务器成功")
	ctx := context.Background()
	for message := range partitionConsumer.Messages() {
		err := json.Unmarshal(message.Value, &order)
		if err != nil {
			log.Println(err)
			logs.WriteLog(err)
			continue
		}
		result := dao.Db.Create(&order)
		key := orderStatusKey(order.UserId, order.ProductId)
		if result.Error != nil {
			log.Println(result.Error)
			logs.WriteLog(result.Error)
			logs.WriteData(order)
			// 写库失败：不能 DEL（否则查询会变成 not_found），标记 failed 并尽力回补库存
			if err := dao.Rdb.Set(ctx, key, model.StatusFailed, orderStatusTTL).Err(); err != nil {
				log.Println("标记订单 failed 失败:", err)
				logs.WriteLog(err)
			}
			if err := dao.Rdb.Incr(ctx, stockKey(order.ProductId)).Err(); err != nil {
				log.Println("回补库存失败:", err)
				logs.WriteLog(err)
			} else {
				log.Printf("写库失败，已标记 failed 并回补库存 product_id=%d user_id=%d\n", order.ProductId, order.UserId)
			}
			continue
		}
		// 落库成功：删缓存，下次查询 miss 回源 MySQL 得到 success
		if err := dao.Rdb.Del(ctx, key).Err(); err != nil {
			log.Println("删除订单状态缓存失败:", err)
			logs.WriteLog(err)
		}
		log.Println("插入成功")
	}
}
