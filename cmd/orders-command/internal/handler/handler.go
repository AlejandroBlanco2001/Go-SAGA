package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"saga-pattern/internal/client"

	"github.com/uptrace/bun"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func StartServer(lc fx.Lifecycle, db *bun.DB, logger *zap.Logger, api client.API) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting server on port 8080")
				if err := http.ListenAndServe(":8080", NewHandler(logger, db, ctx, api)); err != nil {
					logger.Error("Failed to start server", zap.Error(err))
				}
			}()
			return nil
		},
	})
}

func NewHandler(logger *zap.Logger, db *bun.DB, ctx context.Context, api client.API) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Orders service is running"))
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		orders, err := GetOrders(r.Context(), db)

		if err != nil {
			logger.Error("Failed to get orders", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get orders"))
			return
		}

		if len(*orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orders)
	})

	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		order, err := GetOrder(r.Context(), db, id)

		if err != nil {
			logger.Error("Failed to get order", zap.Error(err), zap.String("id", id))
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Order not found"})
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	})

	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		order, err := CreateOrder(r.Context(), db, r, api)

		if err != nil {
			logger.Error("Failed to create order", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to create order"))
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
	})

	return mux
}

var Module = fx.Module("orders-command",
	fx.Invoke(StartServer),
)
