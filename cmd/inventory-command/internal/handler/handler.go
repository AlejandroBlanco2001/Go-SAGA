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
		w.Write([]byte("Inventory service is running"))
	})

	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		inventory, err := GetInventory(r.Context(), db)
		if err != nil {
			logger.Error("Failed to get inventory", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(*inventory) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(inventory)
	})

	mux.HandleFunc("GET /inventory/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		inventory, err := GetInventoryByID(r.Context(), db, id)

		if err != nil {
			logger.Error("Failed to get inventory", zap.Error(err), zap.String("id", id))

			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Inventory not found"})
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(inventory)
	})

	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		inventory, err := CreateInventory(r.Context(), db, r, api)

		if err != nil {
			logger.Error("Failed to create inventory", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(inventory)
	})

	mux.HandleFunc("PUT /inventory/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		inventory, err := UpdateInventory(r.Context(), db, id, r, api)

		if err != nil {
			logger.Error("Failed to update inventory", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(inventory)
	})

	return mux
}

var Module = fx.Module("inventory-command",
	fx.Invoke(StartServer),
)
