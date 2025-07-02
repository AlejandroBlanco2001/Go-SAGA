package main

import (
	"saga-pattern/cmd/inventory-command/internal/handler"
	"saga-pattern/internal/client"
	"saga-pattern/internal/database"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(zap.NewExample),
		client.Module,
		database.Module,
		handler.Module,
	).Run()
}
