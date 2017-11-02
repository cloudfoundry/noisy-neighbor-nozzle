package web

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"code.cloudfoundry.org/noisyneighbor/cache"
	"github.com/gorilla/mux"
)

// TopN is a getter func for getting stats for the top log emitters.
type TopN func(int) []cache.Stat

// Server handles setting up an HTTP server and servicing HTTP requests.
type Server struct {
	lis    net.Listener
	server *http.Server
}

// NewServer opens a TCP listener and returns an initialized Server.
func NewServer(port uint16, username, password string, tn TopN) *Server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to start listener: %d", port)
	}

	log.Printf("Server bound to %s", lis.Addr().String())

	router := mux.NewRouter()

	authMiddleware := BasicAuthMiddleware(username, password)

	router.Handle("/offenders", authMiddleware(OffendersIndex(tn))).
		Methods(http.MethodGet)

	return &Server{
		lis:    lis,
		server: &http.Server{Handler: router},
	}
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
