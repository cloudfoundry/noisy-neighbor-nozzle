package main

import (
	"log"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/datadogreporter"
)

func main() {
	cfg := app.LoadConfig()

	log.Printf("initializing collector with nozzles: %+v", cfg.NozzleAddrs)
	collector := app.NewCollector(cfg.NozzleAddrs)

	log.Printf("initializing datadog reporter")
	reporter := datadogreporter.New(
		cfg.DatadogAPIKey,
		collector,
		datadogreporter.WithHost(cfg.ReporterHost),
		datadogreporter.WithInterval(cfg.ReportInterval),
	)
	reporter.Run()
}
