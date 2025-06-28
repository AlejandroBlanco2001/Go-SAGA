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
	Products string  `json:"products"`
	UserID   int64   `json:"userID"`
}

func GetOrders(ctx context.Context, db *bun.DB) (*[]models.Order, error) {
	orders := new([]models.Order)
	err := db.NewSelect().Model(orders).Limit(20).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func CreateOrder(ctx context.Context, db *bun.DB, r *http.Request) (*models.Order, error) {
	var payload OrderPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	order := &models.Order{
		Price:    payload.Price,
		Products: payload.Products,
		UserID:   payload.UserID,
	}

	if _, err := db.NewInsert().Model(order).Exec(ctx); err != nil {
		return nil, err
	}

	return order, nil
}
