package web

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/store"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// RateStore is the interface from which the server will get rates to be
// rendered via HTTP in JSON.
type RateStore interface {
	Rate(int64) (store.Rate, error)
}

// Server handles setting up an HTTP server and servicing HTTP requests.
type Server struct {
	lis       net.Listener
	server    *http.Server
	logWriter io.Writer
}

// NewServer opens a TCP listener and returns an initialized Server.
func NewServer(
	port uint16,
	ct CheckToken,
	rs RateStore,
	rateInterval time.Duration,
	opts ...ServerOption,
) *Server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to start listener: %d", port)
	}

	log.Printf("Server bound to %s", lis.Addr().String())

	router := mux.NewRouter()

	router.Handle("/rates/{timestamp:[0-9]+}", RatesShow(rs, rateInterval)).
		Methods(http.MethodGet)

	authMiddleware := AdminAuthMiddleware(ct)

	s := &Server{
		lis:       lis,
		logWriter: os.Stdout,
	}

	for _, o := range opts {
		o(s)
	}

	s.server = &http.Server{
		Handler: handlers.LoggingHandler(s.logWriter, authMiddleware(router)),
	}

	return s
}

// Addr returns the address that the listener is bound to.
func (s *Server) Addr() string {
	return s.lis.Addr().String()
}

// Serve serves the HTTP server on the servers Listener. This is a blocking
// method.
func (s *Server) Serve() {
	log.Println(s.server.Serve(s.lis))
}

// Stop will perform a graceful shutdown of the HTTP server.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println(s.server.Shutdown(ctx))
}

// ServerOption is a function that can be passed to the server initializer to
// configure optional settings.
type ServerOption func(*Server)

// WithLogWriter will override the logger used for HTTP logs.
func WithLogWriter(w io.Writer) ServerOption {
	return func(s *Server) {
		s.logWriter = w
	}
}
