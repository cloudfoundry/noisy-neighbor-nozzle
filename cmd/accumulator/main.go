package main

import (
	"code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/accumulator/app"
)

func main() {
	cfg := app.LoadConfig()
	app.New(cfg).Run()
}
