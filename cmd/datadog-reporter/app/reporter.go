package app

import (
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/auth"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/datadog"
)

// Reporter is the constructor for the datadog reporter application.
type Reporter struct {
	reporter *datadog.Reporter
}

// NewReporter configures and returns a new Reporter
func NewReporter(cfg Config) *Reporter {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: cfg.TLSConfig,
		},
	}

	a := auth.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.UAAAddr,
		auth.WithHTTPClient(client),
	)

	httpStore := collector.NewHTTPAppInfoStore(cfg.CAPIAddr, client, a)
	cache := collector.NewCachedAppInfoStore(httpStore)

	log.Printf("initializing collector with accumulator: %+v", cfg.AccumulatorAddr)
	c := collector.New([]string{cfg.AccumulatorAddr}, a, "", cache,
		collector.WithReportLimit(cfg.ReportLimit),
	)

	log.Printf("initializing datadog reporter")
	r := datadog.NewReporter(cfg.DatadogAPIKey, c,
		datadog.WithHost(cfg.ReporterHost),
		datadog.WithInterval(cfg.ReportInterval),
	)

	return &Reporter{
		reporter: r,
	}
}

// Run starts the datadog reporter. This is a blocking method call.
func (r *Reporter) Run() {
	r.reporter.Run()
}
