package main

import "code.cloudfoundry.org/noisyneighbor/internal/app"

func main() {
	cfg := app.LoadConfig()
	nn := app.New(cfg)

	nn.Run()
}
