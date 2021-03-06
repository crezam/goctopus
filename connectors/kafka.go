package connectors

import (
	"github.com/Shopify/sarama"
	"log" // TODO: use logrus
	"time"
)

type KafkaConnector struct {
	KafkaURL   string
	KafkaTopic string
}

//Created by stupendous-man & crezam
func (kc *KafkaConnector) Send(request RequestMessage) {

	producer, err := sarama.NewAsyncProducer([]string{kc.KafkaURL}, nil)

	if err != nil {
		//Retry a few times before going into panic
		for retries := 1; retries <= 5; retries++ {
			producer, err = sarama.NewAsyncProducer([]string{kc.KafkaURL}, nil)

			if err != nil {
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}
	}

	//Deferred function to close producer
	defer func() {
		producer.AsyncClose()
	}()

	// send  a message synchronously when send method is called
	msg := &sarama.ProducerMessage{Topic: kc.KafkaTopic, Value: sarama.StringEncoder(request)}
	producer.Input() <- msg
}

func (kc *KafkaConnector) Receive() ResponseMessage {

	consumer, err := sarama.NewConsumer([]string{kc.KafkaURL}, nil)
	if err != nil {
		//Retry a few times before going into panic
		for retries := 1; retries <= 5; retries++ {
			consumer, err = sarama.NewConsumer([]string{kc.KafkaURL}, nil)

			if err != nil {
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}
	}

	//Deferred function to close consumer
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(kc.KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	//Deferred function to close partition consumer
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// receive one message synchronously on each receive call of the KafkaConnector
	msg := <-partitionConsumer.Messages()

	return ResponseMessage(string(msg.Value))
}
