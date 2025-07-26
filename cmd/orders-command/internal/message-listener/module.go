package message_listener

import "go.uber.org/fx"

var Module = fx.Module("message-listener",
	fx.Invoke(StartKafkaListener),
) 