package app

import (
	"log"
	"time"

	envstruct "code.cloudfoundry.org/go-envstruct"
)

type Config struct {
	NozzleAddrs    []string      `env:"NOZZLE_ADDRS,    required"`
	DatadogAPIKey  string        `env:"DATADOG_API_KEY, required"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ReporterHost   string        `env:"REPORTER_HOST"`
}

func LoadConfig() Config {
	cfg := Config{
		ReportInterval: time.Minute,
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	return cfg
}
