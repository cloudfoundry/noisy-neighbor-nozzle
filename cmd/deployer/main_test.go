package main

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("deploying noisy neighbor via script", func() {
	It("generates a valid manifest", func() {
		manifestPath := WriteManifest(
			input{
				SystemDomain:       "sys-domain.com",
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nn-nozzle",
				AccumulatorAppName: "nn-accumulator",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.sys-domain.com",
				LoggregatorAddr:    "wss://doppler.sys-domain.com:443",
				ClientID:           "noisy-neighbor-client",
				ClientSecret:       "password",
				SkipCertVerify:     false,
			},
		)

		Expect(manifestPath).To(BeAnExistingFile())

		actualManifest, err := ioutil.ReadFile(manifestPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(actualManifest).To(MatchYAML(
			`
				---
				applications:
				  - name: nn-nozzle
				    buildpack: binary_buildpack
				    command: ./noisy-neighbor-nozzle
				    memory: 128M
				    instances: 3
				    env:
					  UAA_ADDR: "https://uaa.sys-domain.com"
				      CLIENT_ID: noisy-neighbor-client
				      CLIENT_SECRET: password
					  LOGGREGATOR_ADDR: "wss://doppler.sys-domain.com:443"
				      SKIP_CERT_VERIFY: false
				  - name: $ACCUMULATOR_APP_NAME
				    buildpack: binary_buildpack
				    command: ./noisy-neighbor-accumulator
				    memory: 128M
				    instances: 1
				    env:
					  UAA_ADDR: "https://uaa.sys-domain.com"
				      CLIENT_ID: noisy-neighbor-client
				      CLIENT_SECRET: password
				      NOZZLE_ADDRS: http://nn-nozzle.app-domain.com
				      NOZZLE_COUNT: 3
				      SKIP_CERT_VERIFY: false
			`,
		))
	})
})
