package client

import (
	"context"
	"os"
	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type KafkaClient struct {
	writer     *kafka.Writer
	reader     *kafka.Reader
	inputChan  MessageChan
	outputChan MessageChan
	ctx        context.Context
	topic      string
}

var topic = os.Getenv("SERVICE_TOPIC")

func NewClient(lc fx.Lifecycle, logger *zap.Logger, inputChan MessageChan, outputChan MessageChan) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"kafka:9092"},
		Balancer: &kafka.LeastBytes{},
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   topic,
	})

	client := &KafkaClient{
		writer:     writer,
		reader:     reader,
		inputChan:  inputChan,
		outputChan: outputChan,
		ctx:        context.Background(),
		topic:      topic,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting to write messages to Kafka", zap.String("topic", client.topic))
				for {
					select {
					case message := <-client.inputChan:
						logger.Info("Writing message to Kafka", zap.Any("message", message), zap.String("topic", client.topic))
						if err := client.writer.WriteMessages(context.Background(), message); err != nil {
							logger.Error("Failed to write message to Kafka", zap.Error(err), zap.String("topic", client.topic))
						} else {
							logger.Info("Successfully wrote message to Kafka", zap.String("topic", client.topic))
						}
					case <-client.ctx.Done():
						return
					}
				}
			}()

			go func() {
				logger.Info("Starting to read messages from Kafka", zap.String("topic", client.topic))
				for {
					message, err := client.reader.ReadMessage(context.Background())
					if err != nil {
						logger.Error("Failed to read message", zap.Error(err), zap.String("topic", client.topic))
						continue
					}
					logger.Info("Read message from Kafka", zap.Any("message", message), zap.String("topic", client.topic))
					client.outputChan <- message
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			client.writer.Close()
			client.reader.Close()
			return nil
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
	fx.Provide(
		fx.Annotate(
			func(params struct {
				fx.In
				Logger     *zap.Logger
				InputChan  MessageChan `name:"inputChan"`
				OutputChan MessageChan `name:"outputChan"`
			}) API {
				return NewAPI(params.Logger, params.InputChan, params.OutputChan)
			},
		),
	),
	fx.Invoke(
		func(params struct {
			fx.In
			Lc         fx.Lifecycle
			Logger     *zap.Logger
			InputChan  MessageChan `name:"inputChan"`
			OutputChan MessageChan `name:"outputChan"`
		}) error {
			return NewClient(params.Lc, params.Logger, params.InputChan, params.OutputChan)
		},
	),
)
