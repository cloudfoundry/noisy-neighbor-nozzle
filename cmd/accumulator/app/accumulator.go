package app

import (
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/auth"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/web"
)

// Accumulator is the constructor for the accumulator application.
type Accumulator struct {
	server *web.Server
}

// New configures and returns a new Accumulator
func New(cfg Config) *Accumulator {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: cfg.TLSConfig,
		},
	}

	a := auth.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.UAAAddr,
		auth.WithHTTPClient(client),
	)

	log.Printf("initializing collector with nozzles: %+v", cfg.NozzleAddrs)
	c := collector.New(cfg.NozzleAddrs, a, cfg.NozzleAppGUID, nil,
		collector.WithHTTPClient(client),
		collector.WithLagerLogger(cfg.MinLogLevel),
	)
	s := web.NewServer(cfg.Port, a.CheckToken, c, cfg.RateInterval,
		web.WithLogWriter(cfg.LogWriter),
	)

	return &Accumulator{
		server: s,
	}
}

// Run starts the accumulator. This is a blocking method call.
func (a *Accumulator) Run() {
	a.server.Serve()
}
