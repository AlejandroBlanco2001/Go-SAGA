package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"saga-pattern/internal/client"
	"saga-pattern/internal/database/models"

	"github.com/segmentio/kafka-go"
	"github.com/uptrace/bun"
)

const (
	InventoryTopic = "inventory"
	InventoryKey   = "InventoryCreated"
)

type InventoryPayload struct {
	Product  string `json:"product"`
	Quantity int64  `json:"quantity"`
}

func GetInventory(ctx context.Context, db *bun.DB) (*[]models.Inventory, error) {
	inventory := new([]models.Inventory)
	err := db.NewSelect().Model(inventory).Limit(20).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return inventory, nil
}

func GetInventoryByID(ctx context.Context, db *bun.DB, id string) (*models.Inventory, error) {
	inventory := new(models.Inventory)

	err := db.NewSelect().Model(inventory).Where("id = ?", id).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return inventory, nil
}

func CreateInventory(ctx context.Context, db *bun.DB, r *http.Request, api client.API) (*models.Inventory, error) {
	var payload InventoryPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	inventory := &models.Inventory{
		ProductID: payload.Product,
		Quantity:  payload.Quantity,
	}

	if _, err := db.NewInsert().Model(inventory).Exec(ctx); err != nil {
		return nil, err
	}

	value := map[string]interface{}{
		"id":       inventory.ID,
		"product":  inventory.ProductID,
		"quantity": inventory.Quantity,
	}

	jsonValue, err := json.Marshal(value)

	if err != nil {
		return nil, err
	}

	message := kafka.Message{
		Topic: InventoryTopic,
		Key:   []byte(InventoryKey),
		Value: jsonValue,
	}

	if err := api.SendMessage(ctx, message); err != nil {
		return nil, err
	}

	return inventory, nil
}

func UpdateInventory(ctx context.Context, db *bun.DB, id string, r *http.Request, api client.API) (*models.Inventory, error) {
	var payload InventoryPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	inventory, err := GetInventoryByID(ctx, db, id)

	if err != nil {
		return nil, err
	}

	inventory.Quantity = payload.Quantity

	if _, err := db.NewUpdate().Model(inventory).Where("id = ?", id).Exec(ctx); err != nil {
		return nil, err
	}

	return inventory, nil
}
