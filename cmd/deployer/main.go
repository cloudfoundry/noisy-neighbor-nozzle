package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/deployer/deploy"
	"code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/deployer/manifest"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var logger = log.New(os.Stderr, "", 0)

func main() {
	in := deploy.Values()
	err := deploy.Validate(in)
	if err != nil {
		logger.Fatalf("Input validation failed: %s", err)
	}

	if in.Interactive {
		in = inputFromUser()
	}

	manifestPath := manifest.Write(in)

	execute("push noisy-neighbor-nozzle", "cf", "push", "-f", manifestPath, "--no-start")
	execute("set subscription ID", "cf", "set-env", in.NozzleAppName, "SUBSCRIPTION_ID", randString(64))

	cmd := exec.Command("cf", "app", in.NozzleAppName, "--guid")
	appGUID, err := cmd.Output()
	if err != nil {
		logger.Fatalf("Failed to get nozzle application GUID: %s", err)
	}
	appGUID = bytes.TrimSpace(appGUID)

	execute("set nozzle app GUID", "cf", "set-env", in.AccumulatorAppName, "NOZZLE_APP_GUID", string(appGUID))
	execute("start nozzle", "cf", "start", in.NozzleAppName)
	execute("start accumulator", "cf", "start", in.AccumulatorAppName)
}

func execute(action string, args ...string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to %s: %s", action, err)
	}
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func inputFromUser() deploy.Input {
	resp := deploy.Input{
		NozzleAppName:      "nn-nozzle",
		AccumulatorAppName: "nn-accumulator",
	}
	qs := []*survey.Question{
		{
			Name: "SystemDomain",
			Prompt: &survey.Input{
				Message: "What is your system domain?",
			},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, &resp)

	qs = []*survey.Question{
		{
			Name: "AppDomain",
			Prompt: &survey.Input{
				Message: "What is your app domain?",
				Default: resp.SystemDomain,
			},
			Validate: survey.Required,
		},
		{
			Name: "UAAAddr",
			Prompt: &survey.Input{
				Message: "UAA Address?",
				Default: fmt.Sprintf("https://uaa.%s", resp.SystemDomain),
			},
			Validate: survey.Required,
		},
		{
			Name: "LoggregatorAddr",
			Prompt: &survey.Input{
				Message: "Loggregator Address?",
				Default: fmt.Sprintf("wss://doppler.%s:443", resp.SystemDomain),
				Help:    "The address of the Cloud Foundry deployed Loggregator. This can be discovered by looking at the doppler_logging_endpoint field from cf curl v2/info.",
			},
			Validate: survey.Required,
		},
		{
			Name: "NozzleInstances",
			Prompt: &survey.Input{
				Message: "How many nozzle instances?",
				Default: "4",
				Help:    "The number of noisy neighbor nozzles to deploy.",
			},
			Validate: unsignedInteger,
		},
		{
			Name: "ClientID",
			Prompt: &survey.Input{
				Message: "Client ID?",
				Help:    "The UAA client ID with the correct authorities.",
			},
			Validate: survey.Required,
		},
		{
			Name: "ClientSecret",
			Prompt: &survey.Password{
				Message: "Client Secret?",
				Help:    "The corresponding secret for the given client ID.",
			},
			Validate: survey.Required,
		},
		{
			Name: "SkipCertVerify",
			Prompt: &survey.Confirm{
				Message: "Skip certificate verification?",
				Help:    "Yes if the Cloud Foundry is using self signed certs.",
			},
		},
	}
	survey.Ask(qs, &resp)

	return resp
}

func unsignedInteger(val interface{}) error {
	vs, ok := val.(string)
	if !ok {
		return errors.New("Couldn't decode type")
	}
	_, err := strconv.ParseUint(vs, 10, 64)
	return err

}
