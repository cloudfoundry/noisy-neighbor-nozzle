package manifest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/deployer/deploy"
)

var logger = log.New(os.Stderr, "", 0)

func Write(in deploy.Input) string {
	manifest, err := ioutil.ReadFile("manifest_template.yml")
	if err != nil {
		logger.Fatalf("Failed to read manifest template: %s", err)
	}

	for k, v := range toEnv(in) {
		manifest = bytes.Replace(manifest, []byte("$"+k), []byte(v), -1)
	}

	outfile, err := ioutil.TempFile("", "manifest-")
	if err != nil {
		logger.Fatalf("Failed to create temporary file: %s", err)
	}
	defer outfile.Close()

	_, err = outfile.Write(manifest)
	if err != nil {
		logger.Fatalf("Failed writing to temporary manifest: %s", err)
	}

	return outfile.Name()
}

func toEnv(i deploy.Input) map[string]string {
	return map[string]string{
		"SYSTEM_DOMAIN":        i.SystemDomain,
		"APP_DOMAIN":           i.AppDomain,
		"NOZZLE_APP_NAME":      i.NozzleAppName,
		"ACCUMULATOR_APP_NAME": i.AccumulatorAppName,
		"NOZZLE_INSTANCES":     fmt.Sprint(i.NozzleInstances),
		"UAA_ADDR":             i.UAAAddr,
		"LOGGREGATOR_ADDR":     i.LoggregatorAddr,
		"CLIENT_ID":            i.ClientID,
		"CLIENT_SECRET":        i.ClientSecret,
		"SKIP_CERT_VERIFY":     fmt.Sprint(i.SkipCertVerify),
	}
}
