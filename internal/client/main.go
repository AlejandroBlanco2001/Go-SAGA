package client

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type KafkaClient struct {
	conn       *kafka.Conn
	inputChan  MessageChan
	outputChan MessageChan
}

func NewClient(lc fx.Lifecycle, logger *zap.Logger, topic string) error {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka:9092", topic, 0)

	if err != nil {
		logger.Error("Failed to dial leader", zap.Error(err))
		panic(err)
	}

	client := &KafkaClient{
		conn:       conn,
		inputChan:  make(MessageChan),
		outputChan: make(MessageChan),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				for {
					select {
					case message := <-client.inputChan:
						client.conn.WriteMessages(message)
					case <-ctx.Done():
						return
					}
				}
			}()

			go func() {
				for {
					message, err := client.conn.ReadMessage(100)
					if err != nil {
						logger.Error("Failed to read message", zap.Error(err))
						continue
					}
					client.outputChan <- message
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	return nil
}

var Module = fx.Options(
	fx.Provide(func() MessageChan {
		return make(MessageChan)
	}),
	fx.Provide(NewAPI),
	fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) error {
		// TODO: Add topic name as an environment variable
		return NewClient(lc, logger, "orders")
	}),
)
