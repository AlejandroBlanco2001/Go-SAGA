package main

import (
	"context"
	"saga-pattern/cmd/orders-command/internal/handler"
	"saga-pattern/cmd/orders-command/internal/message-listener"
	"saga-pattern/internal/client"
	"saga-pattern/internal/database"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var ctx, cancel = context.WithCancel(context.Background())

var options = fx.Options(
	fx.Provide(func() context.Context { return ctx }),
	fx.Provide(zap.NewExample),
	client.Module,
	database.Module,
	handler.Module,
	message_listener.Module,
)

func main() {
	defer cancel()

	fx.New(options).Run()
}
