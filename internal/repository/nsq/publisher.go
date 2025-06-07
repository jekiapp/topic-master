package nsq

import (
	"fmt"

	nsq "github.com/nsqio/go-nsq"
)

// Publish publishes a message to the given topic on all provided nsqd hosts using go-nsq.
func Publish(topic string, message string, host string) error {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(host, config)
	if err != nil {
		return fmt.Errorf("failed to create producer for %s: %w", host, err)
	}
	err = producer.Publish(topic, []byte(message))
	producer.Stop()
	if err != nil {
		return fmt.Errorf("failed to publish to %s: %w", host, err)
	}

	return nil
}
