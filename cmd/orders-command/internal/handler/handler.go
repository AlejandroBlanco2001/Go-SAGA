package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/uptrace/bun"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func StartServer(lc fx.Lifecycle, db *bun.DB, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting server on port 8080")
				if err := http.ListenAndServe(":8080", NewHandler(logger, db, ctx)); err != nil {
					logger.Error("Failed to start server", zap.Error(err))
				}
			}()
			return nil
		},
	})
}

func NewHandler(logger *zap.Logger, db *bun.DB, ctx context.Context) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		orders, err := GetOrders(ctx, db)

		if err != nil {
			logger.Error("Failed to get orders", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get orders"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orders)
	})

	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		order, err := CreateOrder(ctx, db, r)

		if err != nil {
			logger.Error("Failed to create order", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to create order"))
			return
		}

		logger.Info("Order created: ", zap.Any("order", order))

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Order created successfully"))
	})

	return mux
}

var Module = fx.Module("orders-command",
	fx.Invoke(StartServer),
)
