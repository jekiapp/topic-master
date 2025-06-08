package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

type TailMessageInput struct {
	Topic     string   `json:"topic"`
	LimitMsg  int      `json:"limit_msg"`
	NSQDHosts []string `json:"nsqd_hosts"`
}

type TailMessageUsecase struct {
	NSQLookupdAddr string
}

func NewTailMessageUsecase(nsqLookupdAddr string) *TailMessageUsecase {
	return &TailMessageUsecase{
		NSQLookupdAddr: nsqLookupdAddr,
	}
}

// TailAndStream streams up to input.LimitMsg messages to the websocket connection, delimited by ASCII RS (\x1E)
func (u *TailMessageUsecase) HandleTailMessage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	input := TailMessageInput{}
	input.Topic = q.Get("topic")
	limitMsgStr := q.Get("limit_msg")
	if limitMsgStr != "" {
		if n, err := strconv.Atoi(limitMsgStr); err == nil {
			input.LimitMsg = n
		}
	}
	input.NSQDHosts = q["nsqd_hosts"]

	if input.LimitMsg <= 0 {
		http.Error(w, "limit_msg must be > 0", http.StatusBadRequest)
		return
	}
	if input.Topic == "" || len(input.NSQDHosts) == 0 {
		http.Error(w, "topic and nsqd_hosts are required", http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	err = u.tailMessage(r.Context(), conn, input)
	if err != nil {
		log.Println("failed to tail message:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u *TailMessageUsecase) tailMessage(ctx context.Context, conn *websocket.Conn, input TailMessageInput) error {
	const RS = "\x1E"
	var count int32
	msgCh := make(chan *nsq.Message, input.LimitMsg)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Goroutine to detect websocket disconnects
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				cancel()
				return
			}
		}
	}()

	handler := nsq.HandlerFunc(func(message *nsq.Message) error {
		if atomic.LoadInt32(&count) < int32(input.LimitMsg) {
			msgCh <- message
		}
		return nil
	})

	config := nsq.NewConfig()
	channelName := "topic-master-tail-channel-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	consumer, err := nsq.NewConsumer(input.Topic, channelName, config)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	consumer.AddHandler(handler)

	// cleanup function to stop consumer, close msgCh, and delete channel from nsqd
	cleanup := func() {
		consumer.Stop()
		for _, host := range input.NSQDHosts {
			// ensure host is host:4150, convert to host:4151 for HTTP
			httpHost := strings.Replace(host, ":4150", ":4151", 1)
			url := fmt.Sprintf("http://%s/channel/delete?topic=%s&channel=%s", httpHost, input.Topic, channelName)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				log.Printf("[TAIL] failed to create delete channel request: %v", err)
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[TAIL] failed to delete channel: %v", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("[TAIL] failed to delete channel, status: %s", resp.Status)
			}
		}
	}
	defer cleanup()

	// change hosts port to 4150
	hosts := make([]string, len(input.NSQDHosts))
	for i, host := range input.NSQDHosts {
		hosts[i] = strings.Replace(host, ":4151", ":4150", 1)
	}
	err = consumer.ConnectToNSQDs(hosts)
	if err != nil {
		return fmt.Errorf("failed to connect to nsqd: %w", err)
	}

loop:
	for atomic.LoadInt32(&count) < int32(input.LimitMsg) {
		select {
		case <-ctx.Done():
			break loop
		case msg := <-msgCh:
			timestamp := time.Unix(0, msg.Timestamp).Format(time.RFC3339)
			jsonMsg, err := json.Marshal(struct {
				Topic     string `json:"topic"`
				Payload   string `json:"payload"`
				Timestamp string `json:"timestamp"`
			}{
				Topic:     input.Topic,
				Payload:   string(msg.Body),
				Timestamp: timestamp,
			})
			if err != nil {
				return err
			}
			log.Println("[TAIL] sending message:", string(jsonMsg))
			if err := conn.WriteMessage(websocket.TextMessage, append(jsonMsg, RS...)); err != nil {
				return err
			}
			atomic.AddInt32(&count, 1)
		}
	}
	return nil
}
