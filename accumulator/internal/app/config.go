package app

import (
	"crypto/tls"
	"log"
	"time"

	envstruct "code.cloudfoundry.org/go-envstruct"
)

type Config struct {
	UAAAddr        string        `env:"UAA_ADDR,        required"`
	ClientID       string        `env:"CLIENT_ID,       required"`
	ClientSecret   string        `env:"CLIENT_SECRET,   required, noreport"`
	NozzleAddrs    []string      `env:"NOZZLE_ADDRS,    required"`
	DatadogAPIKey  string        `env:"DATADOG_API_KEY, required"`
	SkipCertVerify bool          `env:"SKIP_CERT_VERIFY"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ReporterHost   string        `env:"REPORTER_HOST"`
	TLSConfig      *tls.Config
}

func LoadConfig() Config {
	cfg := Config{
		ReportInterval: time.Minute,
		SkipCertVerify: false,
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	cfg.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.SkipCertVerify}

	return cfg
}
