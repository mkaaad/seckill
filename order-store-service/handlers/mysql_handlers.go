package handlers

import (
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
	for message := range partitionConsumer.Messages() {
		err := json.Unmarshal(message.Value, &order)
		if err != nil {
			log.Println(err)
			logs.WriteLog(err)
			continue
		}
		result := dao.Db.Create(&order)
		if result.Error != nil {
			log.Println(err)
			logs.WriteLog(err)
			logs.WriteData(order)
			continue
		}
		log.Println("插入成功")
	}

}
