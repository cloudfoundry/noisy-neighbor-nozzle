package main

import "code.cloudfoundry.org/noisyneighbor"

func main() {
	cfg := noisyneighbor.LoadConfig()
	nn := noisyneighbor.New(cfg)

	nn.Run()
}
