package main

import (
	"log"
	"net/http"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/authenticator"
)

func main() {
	cfg := app.LoadConfig()

	auth := authenticator.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.UAAAddr,
		authenticator.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: cfg.TLSConfig,
			},
		}),
	)

	log.Printf("initializing collector with nozzles: %+v", cfg.NozzleAddrs)
	collector := app.NewCollector(cfg.NozzleAddrs, auth, cfg.NozzleAppGUID,
		app.WithReportLimit(cfg.ReportLimit),
	)

	log.Printf("initializing datadog reporter")
	reporter := datadogreporter.New(
		cfg.DatadogAPIKey,
		collector,
		datadogreporter.WithHost(cfg.ReporterHost),
		datadogreporter.WithInterval(cfg.ReportInterval),
	)
	reporter.Run()
}
