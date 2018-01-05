package app

import (
	"crypto/tls"
	"log"
	"time"

	envstruct "code.cloudfoundry.org/go-envstruct"
)

// Config stores configuration data for the accumulator.
type Config struct {
	UAAAddr         string        `env:"UAA_ADDR,         required"`
	CAPIAddr        string        `env:"CAPI_ADDR,        required"`
	AccumulatorAddr string        `env:"ACCUMULATOR_ADDR, required"`
	ClientID        string        `env:"CLIENT_ID,        required"`
	ClientSecret    string        `env:"CLIENT_SECRET,    required, noreport"`
	DatadogAPIKey   string        `env:"DATADOG_API_KEY,  required"`
	SkipCertVerify  bool          `env:"SKIP_CERT_VERIFY"`
	ReportInterval  time.Duration `env:"REPORT_INTERVAL"`
	ReporterHost    string        `env:"REPORTER_HOST"`
	ReportLimit     int           `env:"REPORT_LIMIT"`

	AppInfoCacheTTL time.Duration `env:"APP_INFO_CACHE_TTL"`

	TLSConfig *tls.Config
}

// LoadConfig loads the configuration settings from the current environment.
func LoadConfig() Config {
	cfg := Config{
		ReportInterval:  time.Minute,
		ReportLimit:     50,
		SkipCertVerify:  false,
		AppInfoCacheTTL: 150 * time.Second,
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	cfg.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.SkipCertVerify}

	return cfg
}
