package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	nsq "github.com/nsqio/go-nsq"
)

type TopicsResponse struct {
	Topics []string `json:"topics"`
}

func fetchTopics(lookupdHTTPAddr string) ([]string, error) {
	resp, err := http.Get(lookupdHTTPAddr + "/topics")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch topics: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var tr TopicsResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("failed to parse topics JSON: %w", err)
	}
	return tr.Topics, nil
}

func main() {
	lookupdHTTPAddr := os.Getenv("NSQLOOKUPD_HTTP_ADDRESS")
	if lookupdHTTPAddr == "" {
		lookupdHTTPAddr = "http://nsqlookupd:4161"
	}
	lookupdTCPAddr := os.Getenv("NSQLOOKUPD_TCP_ADDRESS")
	if lookupdTCPAddr == "" {
		lookupdTCPAddr = "nsqlookupd:4160"
	}

	topics, err := fetchTopics(lookupdHTTPAddr)
	if err != nil {
		log.Fatalf("Error fetching topics: %v", err)
	}
	if len(topics) == 0 {
		log.Println("No topics found from nsqlookupd.")
		return
	}
	fmt.Printf("Fetched topics: %s\n", strings.Join(topics, ", "))

	config := nsq.NewConfig()
	consumers := []*nsq.Consumer{}

	for _, topic := range topics {
		channel := topic + "_go_consumer"
		consumer, err := nsq.NewConsumer(topic, channel, config)
		if err != nil {
			log.Fatalf("Could not create consumer for topic %s: %v", topic, err)
		}
		consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
			fmt.Printf("[%s] Received message: %s\n", topic, string(message.Body))
			return nil
		}))
		err = consumer.ConnectToNSQLookupd(lookupdHTTPAddr)
		if err != nil {
			log.Fatalf("Could not connect to nsqlookupd for topic %s: %v", topic, err)
		}
		consumers = append(consumers, consumer)
		fmt.Printf("Started consumer for topic: %s (lookupd: %s)\n", topic, lookupdHTTPAddr)
	}

	// Wait for interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Shutting down consumers...")
	for _, c := range consumers {
		c.Stop()
	}
}
