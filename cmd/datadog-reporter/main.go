package main

import "code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/datadog-reporter/app"

func main() {
	cfg := app.LoadConfig()
	app.NewReporter(cfg).Run()
}
