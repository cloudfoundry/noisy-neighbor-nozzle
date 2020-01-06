package app

import (
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/auth"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/datadog"
	"log"
	"net/http"
)

// Reporter is the constructor for the datadog reporter application.
type Reporter struct {
	reporter *datadog.Reporter
}

// NewReporter configures and returns a new Reporter
func NewReporter(cfg Config) *Reporter {
	client := &http.Client{
		Timeout: cfg.CAPIRequestTimeout,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: cfg.TLSConfig,
		},
	}

	ddClient := &http.Client{
		Timeout:   cfg.DatadogRequestTimeout,
		Transport: http.DefaultTransport,
	}

	a := auth.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.UAAAddr,
		auth.WithHTTPClient(client),
	)

	httpStore := collector.NewHTTPAppInfoStore(cfg.CAPIAddr, client, a)
	cache := collector.NewCachedAppInfoStore(
		httpStore,
		collector.WithCacheTTL(cfg.AppInfoCacheTTL),
	)

	log.Printf("initializing collector with accumulator: %+v", cfg.AccumulatorAddr)
	c := collector.New([]string{cfg.AccumulatorAddr}, a, "", cache,
		collector.WithReportLimit(cfg.ReportLimit),
		collector.WithHTTPClient(client),
	)

	log.Printf("initializing datadog reporter")
	r := datadog.NewReporter(cfg.DatadogAPIKey, c,
		datadog.WithHost(cfg.ReporterHost),
		datadog.WithInterval(cfg.ReportInterval),
		datadog.WithHTTPClient(ddClient),
	)

	return &Reporter{
		reporter: r,
	}
}

// Run starts the datadog reporter. This is a blocking method call.
func (r *Reporter) Run() {
	r.reporter.Run()
}
