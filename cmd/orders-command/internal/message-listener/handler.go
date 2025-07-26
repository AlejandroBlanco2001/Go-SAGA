package message_listener

import (
	"context"
	"saga-pattern/internal/client"
	"saga-pattern/internal/database/models"

	"github.com/uptrace/bun"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	OrderRevertedKey = "RevertOrder"
)

type OrderRevertedMessage struct {
	OrderID string `json:"order_id"`
}

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
					case OrderRevertedKey:
						if err := handleOrderReverted(db, logger, message.Value, api); err != nil {
							logger.Error("Failed to handle OrderReverted message", zap.Error(err))
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

func handleOrderReverted(db *bun.DB, logger *zap.Logger, value []byte, api client.API) error {
	// The value is a plain order_id string, not JSON.
	orderID := string(value)

	_, err := db.NewUpdate().Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Set("status = ?", models.OrderStatusCanceled).
		Exec(context.Background())

	if err != nil {
		logger.Error("Failed to revert order", zap.Error(err))
		return err
	}

	logger.Info("Reverted order", zap.String("orderID", orderID))

	return nil
}