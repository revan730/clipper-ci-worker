package src

// Config represents configuration for application
type Config struct {
	// RabbitAddress is used for rabbitmq connection
	RabbitAddress string
	// RabbitQueue name to get jobs from
	RabbitQueue string
	// GCRURL is full url of used GCR registry
	// in form of gcr regional url + project name
	// ex. eu.gcr.io/rocket-science-1488228/
	GCRURL string
	// JsonFile holds path to json file used for docker login
	JSONFile   string
	DBAddr     string
	DB         string
	DBUser     string
	DBPassword string
	CDQueue    string
	Verbose    bool
}
