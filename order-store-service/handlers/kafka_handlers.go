package handlers

import (
	"github.com/IBM/sarama"
)

func ReadMessage() (sarama.PartitionConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		return nil, err
	}

	topic := "write-order-to-mysql"
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}

	return partitionConsumer, nil
}
