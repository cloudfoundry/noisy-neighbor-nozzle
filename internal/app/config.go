package app

import (
	"log"

	envstruct "code.cloudfoundry.org/go-envstruct"
)

// BasicAuthCreds stores configuration for Basic Authentication for the noisey
// neighbor web UI.
type BasicAuthCreds struct {
	Username string `env:"BASIC_AUTH_USERNAME, required"`
	Password string `env:"BASIC_AUTH_PASSWORD, required, noreport"`
}

// Config stores configuration data for the noisy neighbor client.
type Config struct {
	ClientID        string `env:"CLIENT_ID,        required"`
	ClientSecret    string `env:"CLIENT_SECRET,    required, noreport"`
	LoggregatorAddr string `env:"LOGGREGATOR_ADDR, required"`
	Port            uint16 `env:"PORT,             required"`
	SubscriptionID  string `env:"SUBSCRIPTION_ID,  required"`
	BufferSize      int    `env:"BUFFER_SIZE"`
	BasicAuthCreds  BasicAuthCreds
}

// LoadConfig loads the Config from the environment
func LoadConfig() Config {
	cfg := Config{
		BufferSize: 10000,
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config from environment: %s", err)
	}

	return cfg
}
