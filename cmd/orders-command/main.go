package main

import (
	"saga-pattern/internal/database"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(zap.NewExample),
		database.Module,
	).Run()
}
