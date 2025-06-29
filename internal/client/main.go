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
	api        API
}

func NewClient(lc fx.Lifecycle, logger *zap.Logger, api API, topic string) (*KafkaClient, error) {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, 0)

	if err != nil {
		logger.Error("Failed to dial leader", zap.Error(err))
		return nil, err
	}

	client := &KafkaClient{
		conn: conn,
		api:  api,
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

	return client, nil
}

var Module = fx.Module("client",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.As(new(KafkaClient)),
		),
	),
)
