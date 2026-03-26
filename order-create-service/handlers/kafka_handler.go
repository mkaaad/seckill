package handlers

import (
	"encoding/json"
	"order-create/model"

	"github.com/IBM/sarama"
)

func sendMessage(order model.Order) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		return err
	}
	defer producer.Close()

	jsonData, err := json.Marshal(order)
	if err != nil {
		return err
	}
	message := &sarama.ProducerMessage{
		Topic: "write-order-to-mysql",
		Value: sarama.ByteEncoder(jsonData),
	}
	_, _, err = producer.SendMessage(message)
	if err != nil {
		return err
	}
	return nil
}
