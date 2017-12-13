package main

import "code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/nozzle/app"

func main() {
	cfg := app.LoadConfig()
	app.New(cfg).Run()
}
