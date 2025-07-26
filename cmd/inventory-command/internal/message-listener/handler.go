package message_listener

import (
	"context"
	"encoding/json"
	"saga-pattern/internal/client"
	"saga-pattern/internal/database/models"

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

func StartKafkaListener(lc fx.Lifecycle, db *bun.DB, logger *zap.Logger, api client.API) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting Kafka message listener")
				
				// Verify database connection is healthy
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
					case "OrderCreated":
						if err := handleOrderCreated(db, logger, message.Value); err != nil {
							logger.Error("Failed to handle OrderCreated message", zap.Error(err))
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

func handleOrderCreated(db *bun.DB, logger *zap.Logger, value []byte) error {
	var orderMsg OrderMessage
	if err := json.Unmarshal(value, &orderMsg); err != nil {
		return err
	}

	logger.Info("Processing OrderCreated message", 
		zap.String("orderID", orderMsg.OrderID),
		zap.String("product", orderMsg.Product),
		zap.Int64("quantity", orderMsg.Quantity))

	// First, check if the product exists in inventory
	inventory := &models.Inventory{}
	err := db.NewSelect().Model(inventory).
		Where("product_id = ?", orderMsg.Product).
		Scan(context.Background())

	if err != nil {
		logger.Error("Failed to get inventory for product", 
			zap.String("product", orderMsg.Product), 
			zap.Error(err))
		
		// If the product doesn't exist, we might want to create it with 0 quantity
		// or handle this case differently based on business logic
		logger.Warn("Product not found in inventory, creating with 0 quantity", 
			zap.String("product", orderMsg.Product))
		
		inventory = &models.Inventory{
			ProductID: orderMsg.Product,
			Quantity:  0,
		}
		
		// Try to insert the new inventory record
		_, insertErr := db.NewInsert().Model(inventory).Exec(context.Background())
		if insertErr != nil {
			logger.Error("Failed to create inventory record", 
				zap.String("product", orderMsg.Product), 
				zap.Error(insertErr))
			return insertErr
		}
	}

	if inventory.Quantity < orderMsg.Quantity {
		logger.Warn("Insufficient inventory for order", 
			zap.String("orderID", orderMsg.OrderID),
			zap.String("product", orderMsg.Product),
			zap.Int64("requested", orderMsg.Quantity),
			zap.Int64("available", inventory.Quantity))
		return nil
	}

	// Update inventory by reducing the quantity
	inventory.Quantity -= orderMsg.Quantity
	_, err = db.NewUpdate().Model(inventory).
		Where("product_id = ?", orderMsg.Product).
		Exec(context.Background())

	if err != nil {
		logger.Error("Failed to update inventory", zap.Error(err))
		return err
	}

	logger.Info("Successfully processed order and updated inventory", 
		zap.String("orderID", orderMsg.OrderID),
		zap.String("product", orderMsg.Product),
		zap.Int64("quantity", orderMsg.Quantity),
		zap.Int64("remaining", inventory.Quantity))

	return nil
}