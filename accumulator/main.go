package main

import (
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/collector"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/web"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/internal/authenticator"
)

func main() {
	cfg := app.LoadConfig()
	log.Printf("Initializing collector with nozzles: %+v", cfg.NozzleAddrs)

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: cfg.TLSConfig,
		},
	}

	auth := authenticator.NewAuthenticator(
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.UAAAddr,
		authenticator.WithHTTPClient(client),
	)
	httpStore := collector.NewHTTPAppInfoStore(cfg.CAPIAddr, client, auth)
	cache := collector.NewCachedAppInfoStore(httpStore)

	log.Printf("initializing collector with nozzles: %+v", cfg.NozzleAddrs)
	c := collector.New(cfg.NozzleAddrs, auth, cfg.NozzleAppGUID, cache,
		collector.WithReportLimit(cfg.ReportLimit),
	)

	s := web.NewServer(cfg.Port, auth.CheckToken, c, web.WithLogWriter(cfg.LogWriter))
	go s.Serve()

	log.Printf("initializing datadog reporter")
	r := datadogreporter.New(cfg.DatadogAPIKey, c,
		datadogreporter.WithHost(cfg.ReporterHost),
		datadogreporter.WithInterval(cfg.ReportInterval),
	)
	r.Run()
}
