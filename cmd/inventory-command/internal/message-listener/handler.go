package message_listener

import (
	"context"
	"encoding/json"
	"saga-pattern/internal/client"
	"saga-pattern/internal/database/models"
	"github.com/segmentio/kafka-go"

	"github.com/uptrace/bun"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type OrderMessage struct {
	ID       int64   `json:"id"`
	OrderID  string  `json:"order_id"`
	UserID   int64   `json:"user_id"`
	Product  string  `json:"product"`
	Quantity int64   `json:"quantity"`
	Price    float64 `json:"price"`
}

const (
	OrderCreatedKey = "OrderCreated"
	OrderCreatedTopic = "orders"
	RevertOrderTopic = "inventory"
	RevertOrderKey = "RevertOrder"
)

func StartKafkaListener(lc fx.Lifecycle, db *bun.DB, logger *zap.Logger, api client.API) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting Kafka message listener")

				if err := db.PingContext(ctx); err != nil {
					logger.Error("Database connection is not healthy", zap.Error(err))
					return
				}
				logger.Info("Database connection verified")
				
				for {
					message, err := api.ReadMessage(ctx)
					
					if err != nil {
						logger.Error("Failed to read message from Kafka", zap.Error(err))
						continue
					}

					logger.Info("Received message from Kafka", 
						zap.String("topic", message.Topic),
						zap.String("key", string(message.Key)),
						zap.String("key_hex", fmt.Sprintf("%x", message.Key)),
						zap.String("value", string(message.Value)))

					key := string(message.Key)

					switch key {
					case OrderCreatedKey:
						if err := handleOrderCreated(db, logger, message.Value, api); err != nil {
							logger.Error("Failed to handle OrderCreated message, reverting message sent", zap.Error(err))
						}
					default:
						logger.Warn("Unknown message type", zap.String("key", key))
					}
				}
			}()
			return nil
		},
	})
}

func handleOrderCreated(db *bun.DB, logger *zap.Logger, value []byte, api client.API) error {
	var orderMsg OrderMessage
	if err := json.Unmarshal(value, &orderMsg); err != nil {
		return err
	}

	var revertMessage kafka.Message = kafka.Message{
		Topic: RevertOrderTopic,
		Key: []byte(RevertOrderKey),
		Value: []byte(orderMsg.OrderID),
	}

	logger.Info("Processing OrderCreated message", 
		zap.String("orderID", orderMsg.OrderID),
		zap.String("product", orderMsg.Product),
		zap.Int64("quantity", orderMsg.Quantity))

	inventory := &models.Inventory{}
	err := db.NewSelect().Model(inventory).
		Where("product_id = ?", orderMsg.Product).
		Scan(context.Background())

	if err != nil {
		logger.Error("Reverting order, failed to get inventory for product", 
			zap.String("product", orderMsg.Product), 
			zap.Error(err))

		api.SendMessage(context.Background(), revertMessage)

		return nil
	}

	if inventory.Quantity < orderMsg.Quantity {
		logger.Warn("Reverting order, insufficient inventory for order", 
			zap.String("orderID", orderMsg.OrderID),
			zap.String("product", orderMsg.Product),
			zap.Int64("requested", orderMsg.Quantity),
			zap.Int64("available", inventory.Quantity))

		api.SendMessage(context.Background(), revertMessage)

		return nil
	}

	inventory.Quantity -= orderMsg.Quantity

	_, err = db.NewUpdate().Model(inventory).
		Where("product_id = ?", orderMsg.Product).
		Exec(context.Background())

	if err != nil {
		logger.Error("Reverting order, failed to update inventory", zap.Error(err))
		api.SendMessage(context.Background(), revertMessage)
		return err
	}

	logger.Info("Successfully processed order and updated inventory", 
		zap.String("orderID", orderMsg.OrderID),
		zap.String("product", orderMsg.Product),
		zap.Int64("quantity", orderMsg.Quantity),
		zap.Int64("remaining", inventory.Quantity))

	return nil
}