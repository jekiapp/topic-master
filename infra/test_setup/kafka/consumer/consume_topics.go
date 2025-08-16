package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	file, err := os.Open("topics.txt")
	if err != nil {
		fmt.Printf("Failed to open topics.txt: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var topics []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		topic := strings.TrimSpace(scanner.Text())
		if topic != "" {
			topics = append(topics, topic)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading topics.txt: %v\n", err)
		os.Exit(1)
	}

	if len(topics) == 0 {
		fmt.Println("No topics found in topics.txt")
		os.Exit(1)
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
		// "debug":             "consumer,cgrp,topic,fetch",
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %v\n", err)
		os.Exit(1)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics(topics, func(c *kafka.Consumer, e kafka.Event) error {
		switch ev := e.(type) {
		case kafka.AssignedPartitions:
			fmt.Printf("Assigned partitions: %v\n", ev.Partitions)
			c.Assign(ev.Partitions)
		case kafka.RevokedPartitions:
			fmt.Printf("Revoked partitions: %v\n", ev.Partitions)
			c.Unassign()
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to topics: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Listening to topics: %v\n", topics)
	for {
		msg, err := consumer.ReadMessage(1 * time.Second)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.IsTimeout() {
			continue
		} else {
			fmt.Printf("Consumer error: %v\n", err)
		}
	}
}
