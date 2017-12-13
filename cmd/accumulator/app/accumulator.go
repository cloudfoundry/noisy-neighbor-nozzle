package app

import (
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/auth"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/datadog"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/web"
)

type Accumulator struct {
	server   *web.Server
	reporter *datadog.Reporter
}

func New(cfg Config) *Accumulator {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: cfg.TLSConfig,
		},
	}

	a := auth.NewAuthenticator(
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.UAAAddr,
		auth.WithHTTPClient(client),
	)

	httpStore := collector.NewHTTPAppInfoStore(cfg.CAPIAddr, client, a)
	cache := collector.NewCachedAppInfoStore(httpStore)

	log.Printf("initializing collector with nozzles: %+v", cfg.NozzleAddrs)
	c := collector.New(cfg.NozzleAddrs, a, cfg.NozzleAppGUID, cache,
		collector.WithReportLimit(cfg.ReportLimit),
	)

	s := web.NewServer(cfg.Port, a.CheckToken, c, web.WithLogWriter(cfg.LogWriter))
	log.Printf("initializing datadog reporter")
	r := datadog.NewReporter(cfg.DatadogAPIKey, c,
		datadog.WithHost(cfg.ReporterHost),
		datadog.WithInterval(cfg.ReportInterval),
	)

	return &Accumulator{
		server:   s,
		reporter: r,
	}
}

func (a *Accumulator) Run() {
	go a.server.Serve()
	a.reporter.Run()
}
