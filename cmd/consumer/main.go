package consumer

import (
	"transfers-api/internal/clients"
	"transfers-api/internal/config"
	"transfers-api/internal/logging"
)

func main() {

	// init logger
	logger := logging.Logger
	logger.Info("Client logger started")

	// init config
	cfg := config.ParseFromEnv()
	logger.Infof("Client config loaded: %v", cfg.String())

	// init clients
	transfersDBListener := clients.NewClientRabbitMQClient(cfg.RabbitMQConfig)

	transfersDBListener.Listener()

	logger.Info("clients listener created")
}
