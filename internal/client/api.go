package client

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageChan chan kafka.Message

type API interface {
	SendMessage(ctx context.Context, message kafka.Message) error
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

type api struct {
	inputChan  MessageChan
	outputChan MessageChan
	logger     *zap.Logger
}

func NewAPI(logger *zap.Logger, inputChan MessageChan, outputChan MessageChan) API {
	return &api{
		inputChan:  inputChan,
		outputChan: outputChan,
		logger:     logger,
	}
}

func (a *api) SendMessage(ctx context.Context, message kafka.Message) error {
	a.logger.Info("Sending message", zap.Any("message", message))
	select {
	case a.inputChan <- message:
		a.logger.Info("Message sent", zap.Any("message", message))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *api) ReadMessage(ctx context.Context) (kafka.Message, error) {
	a.logger.Info("Reading message")
	return <-a.outputChan, nil
}
