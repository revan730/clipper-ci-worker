package src

// Config represents configuration for application
type Config struct {
	// RabbitAddress is used for rabbitmq connection
	RabbitAddress string
	// RabbitQueue name to get jobs from
	RabbitQueue string
	Verbose     bool
}
