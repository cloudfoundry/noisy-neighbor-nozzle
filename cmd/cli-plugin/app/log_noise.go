package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"text/tabwriter"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/store"

	"code.cloudfoundry.org/cli/plugin"
)

// LogNoise reports the noisiest neighbors for the given accumulator.
func LogNoise(
	conn plugin.CliConnection,
	args []string,
	httpClient HTTPClient,
	appInfoStore AppInfoStore,
	tableWriter io.Writer,
	log Logger,
) {
	appName := "nn-accumulator"

	if len(args) == 1 {
		appName = args[0]
	}

	if len(args) > 1 {
		log.Fatalf("Invalid number of arguments, expected 0 or 1, got %d", len(args))
	}

	app, err := conn.GetApp(appName)
	if err != nil {
		log.Fatalf("%s", err)
	}

	authToken, err := conn.AccessToken()
	if err != nil {
		log.Fatalf("%s", err)
	}

	producers, err := topLogProducers(app, authToken, httpClient)
	if err != nil {
		log.Fatalf("%s", err)
	}

	appInfos, err := fetchAppInfo(producers, appInfoStore)
	if err != nil {
		log.Printf("%s", err)
	}

	tw := tabwriter.NewWriter(tableWriter, 4, 2, 2, ' ', 0)
	fmt.Fprint(tw, "Volume Last Minute\tApp Instance\n")
	for _, item := range producers {
		fmt.Fprintf(
			tw,
			"%d\t%s\n",
			item.count,
			formattedAppInfo(item.appID, appInfos),
		)
	}
	tw.Flush()
}

func topLogProducers(
	app plugin_models.GetAppModel,
	authToken string,
	httpClient HTTPClient,
) (counts, error) {
	if len(app.Routes) < 1 {
		return nil, fmt.Errorf("No routes found for %s", app.Name)
	}

	req, err := http.NewRequest(http.MethodGet, accumulatorEndpoint(app), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to build request to accumulator: %s", err)
	}
	req.Header.Set("Authorization", authToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Failed to get rates from accumulator, expected 200, got %d.",
			resp.StatusCode,
		)
	}

	var rate store.Rate
	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		return nil, fmt.Errorf("Failed to decode accumulator response: %s", err)
	}

	var c counts
	for k, v := range rate.Counts {
		c = append(c, count{
			appID: collector.GUIDIndex(k),
			count: v,
		})
	}

	sort.Sort(sort.Reverse(c))

	if len(c) > 10 {
		c = c[:10]
	}
	return c, nil
}

func accumulatorEndpoint(app plugin_models.GetAppModel) string {
	appRoute := fmt.Sprintf("%s.%s",
		app.Routes[0].Host,
		app.Routes[0].Domain.Name,
	)
	return fmt.Sprintf(
		"https://%s/rates/%d?truncate_timestamp=true",
		appRoute,
		time.Now().Unix(),
	)
}

func fetchAppInfo(
	producers counts,
	appInfoStore AppInfoStore,
) (map[collector.AppGUID]collector.AppInfo, error) {
	var guids []string
	for _, item := range producers {
		guids = append(guids, item.appID.GUID())
	}

	appInfos, err := appInfoStore.Lookup(guids)
	if err != nil {
		return nil, err
	}
	return appInfos, nil
}

func formattedAppInfo(
	appID collector.GUIDIndex,
	appInfos map[collector.AppGUID]collector.AppInfo,
) string {
	appInfo, ok := appInfos[collector.AppGUID(appID.GUID())]
	if !ok {
		return string(appID)
	}
	return fmt.Sprintf("%s.%s.%s/%s",
		appInfo.Org,
		appInfo.Space,
		appInfo.Name,
		appID.Index(),
	)
}

// Logger defines the interface for logging.
type Logger interface {
	Fatalf(format string, args ...interface{})
	Printf(format string, args ...interface{})
}

// HTTPClient defines the interface for making HTTP requests.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// AppInfoStore defines the interface used to look up app, space and org names.
type AppInfoStore interface {
	Lookup(guids []string) (map[collector.AppGUID]collector.AppInfo, error)
}

type count struct {
	appID collector.GUIDIndex
	count uint64
}

type counts []count

func (c counts) Len() int           { return len(c) }
func (c counts) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c counts) Less(i, j int) bool { return c[i].count < c[j].count }
