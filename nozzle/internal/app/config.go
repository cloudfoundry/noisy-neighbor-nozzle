package app

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	envstruct "code.cloudfoundry.org/go-envstruct"
)

// Config stores configuration data for the noisy neighbor client.
type Config struct {
	UAAAddr           string        `env:"UAA_ADDR,         required"`
	ClientID          string        `env:"CLIENT_ID,        required"`
	ClientSecret      string        `env:"CLIENT_SECRET,    required, noreport"`
	LoggregatorAddr   string        `env:"LOGGREGATOR_ADDR, required"`
	Port              uint16        `env:"PORT,             required"`
	SubscriptionID    string        `env:"SUBSCRIPTION_ID,  required"`
	SkipCertVerify    bool          `env:"SKIP_CERT_VERIFY"`
	BufferSize        int           `env:"BUFFER_SIZE"`
	PollingInterval   time.Duration `env:"POLLING_INTERVAL"`
	MaxRateBuckets    int           `env:"MAX_RATE_BUCKETS"`
	IncludeRouterLogs bool          `env:"INCLUDE_ROUTER_LOGS"`

	// VCapApplication is used to detect whether or not the application is
	// deployed as a CF application.
	VCapApplication string `env:"VCAP_APPLICATION"`
	TLSConfig       *tls.Config

	LogWriter io.Writer
}

// LoadConfig loads the Config from the environment
func LoadConfig() Config {
	cfg := Config{
		SkipCertVerify:    false,
		BufferSize:        10000,
		PollingInterval:   time.Minute,
		MaxRateBuckets:    60,
		IncludeRouterLogs: false,
		LogWriter:         os.Stdout,
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config from environment: %s", err)
	}

	if cfg.VCapApplication != "" {
		cfg.LogWriter = ioutil.Discard
	}

	cfg.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.SkipCertVerify}

	return cfg
}
