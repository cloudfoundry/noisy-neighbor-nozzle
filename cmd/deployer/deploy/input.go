package deploy

import (
	"errors"
	"flag"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Input struct {
	SystemDomain       string
	AppDomain          string
	NozzleAppName      string
	AccumulatorAppName string
	NozzleInstances    uint
	UAAAddr            string
	CAPIAddr           string
	LoggregatorAddr    string
	ClientID           string
	ClientSecret       string
	SkipCertVerify     bool
	Interactive        bool

	DataDogForwarderName string
	DataDogForwarder     bool
	DataDogAPIKey        string
}

func Values() Input {
	sysDomain := flag.String("system-domain", "", "Your system domain")
	appDomain := flag.String("app-domain", "", "Your app domain")
	nozzleAppName := flag.String("nozzle-app-name", "nn-nozzle", "")
	accAppName := flag.String("accumulator-app-name", "nn-accumulator", "")
	dataDogAppName := flag.String("datadog-app-name", "nn-datadog-forwarder", "")
	interactive := flag.Bool(
		"interactive",
		false,
		"True if you want user prompts to configure app setup. Otherwise, will assume all input via flags.",
	)
	instances := flag.Uint(
		"nozzle-instances",
		4,
		"The number of noisy neighbor nozzles to deploy",
	)
	uaa := flag.String(
		"uaa-addr",
		"",
		"In the form https://uaa.<system-domain>. The address of the UAA server in Cloud Foundry. This can be discovered by looking at the token_endpoint field from cf curl v2/info.",
	)
	loggrAddr := flag.String(
		"loggregator-addr",
		"",
		"In the form wss://doppler.<system-domain>:443. The address of the Cloud Foundry deployed Loggregator. This can be discovered by looking at the doppler_logging_endpoint field from cf curl v2/info.",
	)
	clientID := flag.String(
		"client-id",
		"",
		"The UAA client ID with the correct authorities.",
	)
	secret := flag.String(
		"client-secret",
		"",
		"The corresponding secret for the given client ID.",
	)
	skipCertVerify := flag.Bool(
		"skip-cert-verify",
		false,
		"Yes if the Cloud Foundry is using self signed certs.",
	)
	dataDogAPIKey := flag.String(
		"datadog-api-key",
		"",
		"DataDog API key",
	)
	capiAddr := flag.String(
		"capi-addr",
		"",
		"CAPI Address",
	)

	flag.Parse()

	return Input{
		Interactive:          *interactive,
		SystemDomain:         *sysDomain,
		AppDomain:            *appDomain,
		NozzleAppName:        *nozzleAppName,
		AccumulatorAppName:   *accAppName,
		DataDogForwarderName: *dataDogAppName,
		NozzleInstances:      *instances,
		UAAAddr:              *uaa,
		LoggregatorAddr:      *loggrAddr,
		ClientID:             *clientID,
		ClientSecret:         *secret,
		SkipCertVerify:       *skipCertVerify,
		DataDogAPIKey:        *dataDogAPIKey,
		CAPIAddr:             *capiAddr,
	}
}

func Validate(f Input) error {
	if f.Interactive {
		return nil
	}

	if f.SystemDomain == "" {
		return errors.New("System Domain is required, but not set")
	}
	if f.AppDomain == "" {
		return errors.New("App Domain is required, but not set")
	}
	if f.NozzleAppName == "" {
		return errors.New("Nozzle App Name required, but not set")
	}
	if f.NozzleInstances == 0 {
		return errors.New("Nozzle instances count must be greater than 0")
	}
	if f.AccumulatorAppName == "" {
		return errors.New("Accumulator App Name is required, but not set")
	}
	if f.UAAAddr == "" {
		return errors.New("UAA Address is required, but not set")
	}
	if f.LoggregatorAddr == "" {
		return errors.New("Loggrgator Address is required, but not set")
	}
	if f.ClientID == "" {
		return errors.New("Client ID is required, but not set")
	}
	if f.ClientSecret == "" {
		return errors.New("Client Secret is required, but not set")
	}
	return nil

}
