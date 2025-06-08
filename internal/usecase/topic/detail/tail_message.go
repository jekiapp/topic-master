package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

type TailMessageUsecase struct{}

// TailAndStream streams up to input.LimitMsg messages to the websocket connection, delimited by ASCII RS (\x1E)
func (u *TailMessageUsecase) HandleTailMessage(w http.ResponseWriter, r *http.Request) {
	input := TailMessageInput{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return
	}

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u *TailMessageUsecase) tailMessage(ctx context.Context, conn *websocket.Conn, input TailMessageInput) error {
	const RS = "\x1E"
	var count int32
	msgCh := make(chan *nsq.Message, input.LimitMsg)
	handler := nsq.HandlerFunc(func(message *nsq.Message) error {
		if atomic.LoadInt32(&count) < int32(input.LimitMsg) {
			msgCh <- message
		}
		return nil
	})

	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(input.Topic, "topic-master-tail-channel-"+strconv.FormatInt(time.Now().UnixNano(), 10), config)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	err = consumer.ConnectToNSQDs(input.NSQDHosts)
	if err != nil {
		return fmt.Errorf("failed to connect to nsqd: %w", err)
	}

	consumer.AddHandler(handler)
	defer consumer.Stop()

loop:
	for atomic.LoadInt32(&count) < int32(input.LimitMsg) {
		select {
		case <-ctx.Done():
			break loop
		case msg := <-msgCh:
			jsonMsg, err := json.Marshal(struct {
				Topic   string `json:"topic"`
				Payload string `json:"payload"`
			}{
				Topic:   input.Topic,
				Payload: string(msg.Body),
			})
			if err != nil {
				return err
			}
			if err := conn.WriteMessage(websocket.TextMessage, append(jsonMsg, RS...)); err != nil {
				return err
			}
			atomic.AddInt32(&count, 1)
		}
	}
	return nil
}
