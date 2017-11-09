package main

import "code.cloudfoundry.org/noisy-neighbor-nozzle/nozzle/internal/app"

func main() {
	cfg := app.LoadConfig()
	nn := app.New(cfg)

	nn.Run()
}
