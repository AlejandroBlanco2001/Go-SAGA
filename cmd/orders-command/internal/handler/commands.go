package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"saga-pattern/internal/database/models"

	"github.com/uptrace/bun"
)

type OrderPayload struct {
	Price    float64 `json:"price"`
	Product  string  `json:"product"`
	Quantity int64   `json:"quantity"`
	UserID   int64   `json:"user_id"`
}

func GetOrders(ctx context.Context, db *bun.DB) (*[]models.Order, error) {
	orders := new([]models.Order)
	err := db.NewSelect().Model(orders).Limit(20).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func GetOrder(ctx context.Context, db *bun.DB, id string) (*models.Order, error) {
	order := new(models.Order)

	err := db.NewSelect().Model(order).Where("id = ?", id).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return order, nil
}

func CreateOrder(ctx context.Context, db *bun.DB, r *http.Request) (*models.Order, error) {
	var payload OrderPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	order := &models.Order{
		Price:    payload.Price,
		Product:  payload.Product,
		Quantity: payload.Quantity,
		UserID:   payload.UserID,
	}

	if _, err := db.NewInsert().Model(order).Exec(ctx); err != nil {
		return nil, err
	}

	return order, nil
}
