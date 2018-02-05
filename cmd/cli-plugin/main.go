package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/cli-plugin/app"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
)

// LogNoiseCLI represent the CF CLI log-noise plugin
type LogNoiseCLI struct{}

// Run implements the log-noise command
func (c *LogNoiseCLI) Run(conn plugin.CliConnection, args []string) {
	if len(args) == 0 {
		log.Fatalf("Expected at least 1 argument, but got 0.")
	}

	skipSSL, err := conn.IsSSLDisabled()
	if err != nil {
		log.Fatalf("%s", err)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: skipSSL,
	}
	apiEndpoint, err := conn.ApiEndpoint()
	if err != nil {
		log.Fatalf("%s", err)
	}

	httpAppInfoStore := collector.NewHTTPAppInfoStore(
		apiEndpoint,
		http.DefaultClient,
		&Authenticator{conn: conn},
	)

	switch args[0] {
	case "log-noise":
		app.LogNoise(
			conn,
			args[1:],
			http.DefaultClient,
			httpAppInfoStore,
			os.Stdout,
			log.New(os.Stdout, "", 0),
		)
		return
	}
}

// GetMetadata provides usage information for the log-noise command
func (c *LogNoiseCLI) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Log Noise CLI Plugin",
		Commands: []plugin.Command{
			{
				Name: "log-noise",
				UsageDetails: plugin.Usage{
					Usage: "log-noise <nozzle accumulator app name>",
				},
				HelpText: "Show top log producers from noisy-neighbor-nozzle accumulator.",
			},
		},
	}
}

func main() {
	plugin.Start(&LogNoiseCLI{})
}

// Authenticator is used to refresh the authentication token.
type Authenticator struct {
	conn plugin.CliConnection
}

// RefreshAuthToken returns the CAPI auth token without the bearer prefix
func (a *Authenticator) RefreshAuthToken() (string, error) {
	token, err := a.conn.AccessToken()
	if err != nil {
		return "", err
	}

	return strings.Replace(token, "bearer ", "", 1), nil
}
