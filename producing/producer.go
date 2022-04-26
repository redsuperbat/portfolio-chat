package producing

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"rsb.asuscomm.com/portfolio-chat/constants"
)

const (
	ceilRetryTime float64 = 25_000
)

func getRetryTime(numberOfRetries int64) float64 {
	if numberOfRetries > 5 {
		return ceilRetryTime
	}
	retryTime := math.Exp(float64(numberOfRetries))
	return retryTime * 100
}

func NewConn() (func(value []byte) error, func() error) {
	kafkaBroker := os.Getenv(constants.KAFKA_BROKER)
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	kafkaTopic := os.Getenv(constants.KAFKA_TOPIC)
	if kafkaTopic == "" {
		log.Fatalln("A topic must be provided")
	}

	log.Printf("Connecting to broker %s ...", kafkaBroker)
	var err error
	var numberOfRetries int64 = 0
	for err == nil {

		conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaBroker, kafkaTopic, 0)

		if err != nil {
			numberOfRetries += 1
			retryTime := getRetryTime(numberOfRetries)
			log.Printf("failed to dial leader: %s. Retrying in %v seconds", err.Error(), retryTime/1000)
			time.Sleep(time.Duration(retryTime) * time.Millisecond)
			continue
		}

		log.Println("Connection successful!")

		return func(value []byte) error {
			msg := kafka.Message{Value: value}
			if _, err := conn.WriteMessages(msg); err != nil {
				return err
			}
			return nil
		}, conn.Close

	}
	log.Fatalln("Unable to connect to kafka, please try again..")
	return nil, nil
}
