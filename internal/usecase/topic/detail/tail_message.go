package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

// TailMessageInput holds the parameters for tailing messages from NSQ.
// NSQDHosts: list of nsqd TCP endpoints (host:port) to connect to.
// LimitMsg: maximum number of messages to stream.
// Topic: NSQ topic to consume from.
type TailMessageInput struct {
	Topic     string   `json:"topic"`
	LimitMsg  int      `json:"limit_msg"`
	NSQDHosts []string `json:"nsqd_hosts"`
}

// activeChannel tracks the nsqd hosts and topic for a registered channel.
// Motivation: Used for robust cleanup of all active channels on shutdown.
type activeChannel struct {
	nsqdHosts []string
	topic     string
}

// TailMessageUsecase manages the lifecycle of tailing channels and their cleanup.
// Motivation: Tracks all active channels for safe shutdown, prevents new registrations during shutdown, and ensures concurrency safety.
type TailMessageUsecase struct {
	activeChannels map[string]activeChannel // Tracks all active channels for cleanup
	mu             sync.Mutex               // Protects access to activeChannels
	stopping       atomic.Bool              // Set to true when shutdown is initiated
}

// NewTailMessageUsecase creates a new usecase instance and starts a goroutine to listen for OS termination signals.
// Motivation: Ensures all active channels are cleaned up on process exit, and prevents new registrations after shutdown is triggered.
func NewTailMessageUsecase() *TailMessageUsecase {
	u := &TailMessageUsecase{
		activeChannels: make(map[string]activeChannel),
	}
	// Listen for OS signals (SIGINT, SIGTERM) to trigger cleanup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("[TAIL] received termination signal")
		// Set stopping flag to prevent new channel registrations
		u.stopping.Store(true)
		// Copy activeChannels for cleanup outside the lock
		u.mu.Lock()
		channels := make(map[string]activeChannel, len(u.activeChannels))
		for ch, ac := range u.activeChannels {
			channels[ch] = ac
		}
		u.mu.Unlock()
		// Delete all active channels from all nsqd hosts
		for channelName, ac := range channels {
			u.deleteChannelFromNSQDs(ac.topic, channelName, ac.nsqdHosts)
		}
		os.Exit(0)
	}()
	return u
}

// HandleTailMessage upgrades the HTTP connection to a websocket and starts streaming messages.
// Motivation: Provides a websocket endpoint for clients to tail NSQ messages in real time.
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

	err = u.tailMessage(r.Context(), conn, input, nil)
	if err != nil {
		log.Println("failed to tail message:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// tailMessage streams up to input.LimitMsg messages from NSQ to the websocket connection.
// Motivation: Handles message consumption, client disconnects, and resource cleanup efficiently and safely.
func (u *TailMessageUsecase) tailMessage(ctx context.Context, conn *websocket.Conn, input TailMessageInput, signalCh <-chan os.Signal) error {
	const RS = "\x1E" // ASCII Record Separator for message framing
	var count int32
	msgCh := make(chan *nsq.Message, input.LimitMsg)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Goroutine to detect websocket disconnects and cancel context
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				cancel()
				return
			}
		}
	}()

	// NSQ handler: delivers messages to msgCh up to the limit
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

	// Prevent new channel registration if service is stopping
	if u.stopping.Load() {
		return fmt.Errorf("service is stopping, no new channel registrations allowed")
	}

	// Register the active channel for later cleanup
	u.mu.Lock()
	u.activeChannels[channelName] = activeChannel{
		nsqdHosts: input.NSQDHosts,
		topic:     input.Topic,
	}
	u.mu.Unlock()

	// cleanup: stop consumer, delete channel from nsqd, and unregister
	cleanup := func() {
		consumer.Stop()
		u.mu.Lock()
		ac := u.activeChannels[channelName]
		// delete the channel from the nsqd
		u.deleteChannelFromNSQDs(ac.topic, channelName, ac.nsqdHosts)
		// delete the channel from the active channels map
		delete(u.activeChannels, channelName)
		u.mu.Unlock()
	}
	defer cleanup()

	// Prepare NSQD TCP hosts for connection (convert :4151 to :4150)
	hosts := make([]string, len(input.NSQDHosts))
	for i, host := range input.NSQDHosts {
		hosts[i] = strings.Replace(host, ":4151", ":4150", 1)
	}
	err = consumer.ConnectToNSQDs(hosts)
	if err != nil {
		return fmt.Errorf("failed to connect to nsqd: %w", err)
	}

	// Main loop: stream messages, handle context/signal/counter
loop:
	for atomic.LoadInt32(&count) < int32(input.LimitMsg) {
		select {
		case <-ctx.Done():
			break loop
		case <-signalCh:
			log.Println("[TAIL] received termination signal")
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

// deleteChannelFromNSQDs deletes the given channel for the topic from all provided nsqd hosts.
func (u *TailMessageUsecase) deleteChannelFromNSQDs(topic, channelName string, nsqdHosts []string) {
	for _, host := range nsqdHosts {
		httpHost := strings.Replace(host, ":4150", ":4151", 1)
		url := fmt.Sprintf("http://%s/channel/delete?topic=%s&channel=%s", httpHost, topic, channelName)
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
