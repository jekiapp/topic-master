package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
)

type NsqHandlerResult struct {
	Requeue time.Duration
	Finish  bool
}

type GenericHandlerNsq[I any] interface {
	HandleMessage(ctx context.Context, input I) (output NsqHandlerResult, err error)
}

func NsqGenericHandler[I any](handler GenericHandlerNsq[I]) nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		body := msg.Body
		data := new(I)
		if err := json.Unmarshal(body, data); err != nil {
			return fmt.Errorf("error unmarshal object %+v", data)
		}

		ctx := context.Background()

		// Validate input object using json validator
		if err := validate.Struct(data); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		output, err := handler.HandleMessage(ctx, *data)
		if err != nil {
			if output.Requeue != 0 {
				msg.Requeue(output.Requeue)
			} else if output.Finish {
				msg.Finish()
			}

			return err
		}

		msg.Finish()
		return nil
	}
}

type Consumer struct {
	Topic   string
	Channel string
	Config  *nsq.Config
	Handler nsq.Handler
}

func NewGenericConsumer[T any](topic, channel string, config *nsq.Config, handler GenericHandlerNsq[T]) Consumer {
	return Consumer{
		Topic:   topic,
		Channel: channel,
		Config:  config,
		Handler: NsqGenericHandler(handler),
	}
}
