package main

import (
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/authenticator"
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
	httpStore := app.NewHTTPAppInfoStore(cfg.CAPIAddr, client, auth)

	cache := app.NewCachedAppInfoStore(httpStore)

	collector := app.NewCollector(
		cfg.NozzleAddrs,
		auth,
		cfg.NozzleAppGUID,
		cache,
		app.WithReportLimit(cfg.ReportLimit),
	)
	reporter := datadogreporter.New(
		cfg.DatadogAPIKey,
		collector,
		datadogreporter.WithHost(cfg.ReporterHost),
		datadogreporter.WithInterval(cfg.ReportInterval),
	)

	reporter.Run()
}
