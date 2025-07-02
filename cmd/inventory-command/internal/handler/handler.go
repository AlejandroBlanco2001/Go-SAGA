package handler

import (
	"context"
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

	return mux
}

var Module = fx.Module("inventory-command",
	fx.Invoke(StartServer),
)
