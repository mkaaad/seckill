package handlers

import (
	"encoding/json"
	"fmt"
	"order-create/model"
	"sync"

	"github.com/IBM/sarama"
)

var (
	producer     sarama.SyncProducer
	producerOnce sync.Once
	producerErr  error
)

func getProducer() (sarama.SyncProducer, error) {
	producerOnce.Do(func() {
		config := sarama.NewConfig()
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
		config.Producer.Return.Successes = true
		config.Producer.Retry.Max = 3
		producer, producerErr = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	})
	return producer, producerErr
}

// CloseProducer closes the shared Kafka producer (optional, e.g. on process exit).
func CloseProducer() {
	if producer != nil {
		_ = producer.Close()
	}
}

func sendMessage(order model.Order) error {
	p, err := getProducer()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(order)
	if err != nil {
		return err
	}
	// Key by product so same product lands on same partition (ordering within product)
	message := &sarama.ProducerMessage{
		Topic: "write-order-to-mysql",
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", order.ProductId)),
		Value: sarama.ByteEncoder(jsonData),
	}
	_, _, err = p.SendMessage(message)
	return err
}
