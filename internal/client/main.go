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
	ctx        context.Context
}

func NewClient(lc fx.Lifecycle, logger *zap.Logger, inputChan MessageChan, outputChan MessageChan, topic string) error {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka:9092", topic, 0)

	if err != nil {
		logger.Error("Failed to dial leader", zap.Error(err))
		return err
	}

	client := &KafkaClient{
		conn:       conn,
		inputChan:  inputChan,
		outputChan: outputChan,
		ctx:        context.Background(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting to write messages to Kafka")
				for {
					select {
					case message := <-client.inputChan:
						logger.Info("Writing message to Kafka", zap.Any("message", message))
						if _, err := client.conn.WriteMessages(message); err != nil {
							logger.Error("Failed to write message to Kafka", zap.Error(err))
						} else {
							logger.Info("Successfully wrote message to Kafka")
						}
					case <-client.ctx.Done():
						return
					}
				}
			}()

			go func() {
				logger.Info("Starting to read messages from Kafka")
				for {
					message, err := client.conn.ReadMessage(100)
					if err != nil {
						logger.Error("Failed to read message", zap.Error(err))
						continue
					}
					logger.Info("Read message from Kafka", zap.Any("message", message))
					client.outputChan <- message
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.conn.Close()
		},
	})

	return nil
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			func() MessageChan {
				return make(MessageChan)
			},
			fx.ResultTags(`name:"inputChan"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			func() MessageChan {
				return make(MessageChan)
			},
			fx.ResultTags(`name:"outputChan"`),
		),
	),
	fx.Provide(fx.Annotate(
		func(params struct {
			fx.In
			Logger     *zap.Logger
			InputChan  MessageChan `name:"inputChan"`
			OutputChan MessageChan `name:"outputChan"`
		}) API {
			return NewAPI(params.Logger, params.InputChan, params.OutputChan)
		},
	)),
	fx.Invoke(func(params struct {
		fx.In
		Lc         fx.Lifecycle
		Logger     *zap.Logger
		InputChan  MessageChan `name:"inputChan"`
		OutputChan MessageChan `name:"outputChan"`
	}) error {
		// TODO: Add topic name as an environment variable
		return NewClient(params.Lc, params.Logger, params.InputChan, params.OutputChan, "orders")
	}),
)
