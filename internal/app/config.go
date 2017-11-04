package app

import (
	"crypto/tls"
	"log"
	"time"

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
	UAAAddr         string        `env:"UAA_ADDR,         required"`
	ClientID        string        `env:"CLIENT_ID,        required"`
	ClientSecret    string        `env:"CLIENT_SECRET,    required, noreport"`
	LoggregatorAddr string        `env:"LOGGREGATOR_ADDR, required"`
	Port            uint16        `env:"PORT,             required"`
	SubscriptionID  string        `env:"SUBSCRIPTION_ID,  required"`
	SkipCertVerify  bool          `env:"SKIP_CERT_VERIFY"`
	BufferSize      int           `env:"BUFFER_SIZE"`
	PollingInterval time.Duration `env:"POLLING_INTERVAL"`
	BasicAuthCreds  BasicAuthCreds
	TLSConfig       *tls.Config
}

// LoadConfig loads the Config from the environment
func LoadConfig() Config {
	cfg := Config{
		SkipCertVerify:  false,
		BufferSize:      10000,
		PollingInterval: time.Minute,
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config from environment: %s", err)
	}

	cfg.TLSConfig = &tls.Config{InsecureSkipVerify: !cfg.SkipCertVerify}

	return cfg
}
