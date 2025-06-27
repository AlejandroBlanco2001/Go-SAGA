package main

import (
	"saga-pattern/cmd/orders-command/internal/handler"
	"saga-pattern/internal/database"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(zap.NewExample),
		database.Module,
		handler.Module,
	).Run()
}
