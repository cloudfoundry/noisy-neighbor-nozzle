package app

import (
	"log"
	"net/http"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/auth"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/ingress"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/store"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/web"
	"github.com/cloudfoundry/noaa/consumer"
)

// Nozzle is the top level data structure for the Nozzle
// application.
type Nozzle struct {
	cfg        Config
	server     *web.Server
	ingestor   *ingress.Ingestor
	processor  *ingress.Processor
	aggregator *store.Aggregator
}

// New returns an initialized NoisyNeighbor. This will authenticate with UAA,
// open a connection to the firehose, and initialize all subprocesses.
func New(cfg Config) *Nozzle {
	authenticator := auth.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.UAAAddr,
		auth.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: cfg.TLSConfig,
			},
		}),
	)
	token, err := authenticator.RefreshAuthToken()
	if err != nil {
		log.Fatalf("failed to authenticate: %s", err)
	}

	cnsmr := consumer.New(cfg.LoggregatorAddr, cfg.TLSConfig, nil)
	cnsmr.RefreshTokenFrom(authenticator)
	msgs, errs := cnsmr.FilteredFirehose(
		cfg.SubscriptionID,
		token,
		consumer.LogMessages,
	)
	go func() {
		for err := range errs {
			log.Printf("error received from firehose: %s", err)
		}
	}()

	b := ingress.NewBuffer(cfg.BufferSize)
	c := store.NewCounter()
	a := store.NewAggregator(c,
		store.WithPollingInterval(cfg.PollingInterval),
		store.WithMaxRateBuckets(cfg.MaxRateBuckets),
	)
	s := web.NewServer(
		cfg.Port,
		authenticator.CheckToken,
		a,
		cfg.PollingInterval,
		web.WithLogWriter(cfg.LogWriter),
	)

	return &Nozzle{
		cfg:        cfg,
		server:     s,
		aggregator: a,
		ingestor:   ingress.NewIngestor(msgs, b.Set),
		processor:  ingress.NewProcessor(b.Next, c.Inc, cfg.IncludeRouterLogs),
	}
}

// Addr returns the address that the NoisyNeighbor is bound to.
func (n *Nozzle) Addr() string {
	return n.server.Addr()
}

// Run starts the NoisyNeighbor application. This is a blocking method call.
func (n *Nozzle) Run() {
	go n.ingestor.Run()
	go n.processor.Run()
	go n.aggregator.Run()

	n.server.Serve()
}

// Stop gracefully stops the NoisyNeighbor application. It will disconnect from
// the firehose, and complete any active HTTP requests.
func (n *Nozzle) Stop() {
	n.server.Stop()
}
