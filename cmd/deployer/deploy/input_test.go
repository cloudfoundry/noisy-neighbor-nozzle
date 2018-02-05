package deploy_test

import (
	"code.cloudfoundry.org/noisy-neighbor-nozzle/cmd/deployer/deploy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployflags", func() {

	It("doesn't validate when interactive is true", func() {
		f := deploy.Input{
			Interactive: true,
		}
		err := deploy.Validate(f)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("interactive is false",
		func(f deploy.Input) {
			f.Interactive = false
			err := deploy.Validate(f)
			Expect(err).To(HaveOccurred())
		},
		Entry("SystemDomain flag is not set",
			deploy.Input{
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.my-env.com",
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientID:           "my-client",
				ClientSecret:       "a-password",
			},
		),
		Entry("AppDomain flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.my-env.com",
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientID:           "my-client",
				ClientSecret:       "a-password",
			},
		),
		Entry("NozzleAppName flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				AppDomain:          "app-domain.com",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.my-env.com",
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientID:           "my-client",
				ClientSecret:       "a-password",
			},
		),
		Entry("AccumulatorAppName flag is not set",
			deploy.Input{
				SystemDomain:    "system-domain.com",
				AppDomain:       "app-domain.com",
				NozzleAppName:   "nozzle-app",
				NozzleInstances: 3,
				UAAAddr:         "https://uaa.my-env.com",
				LoggregatorAddr: "wss://doppler.my-env.com:443",
				ClientID:        "my-client",
				ClientSecret:    "a-password",
			},
		),
		Entry("NozzleInstances flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				UAAAddr:            "https://uaa.my-env.com",
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientID:           "my-client",
				ClientSecret:       "a-password",
			},
		),
		Entry("UAAAddr flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientID:           "my-client",
				ClientSecret:       "a-password",
			},
		),
		Entry("LoggregatorAddr flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.my-env.com",
				ClientID:           "my-client",
				ClientSecret:       "a-password",
			},
		),
		Entry("ClientID flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.my-env.com",
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientSecret:       "a-password",
			},
		),
		Entry("ClientSecret flag is not set",
			deploy.Input{
				SystemDomain:       "system-domain.com",
				AppDomain:          "app-domain.com",
				NozzleAppName:      "nozzle-app",
				AccumulatorAppName: "accumulator-app",
				NozzleInstances:    3,
				UAAAddr:            "https://uaa.my-env.com",
				LoggregatorAddr:    "wss://doppler.my-env.com:443",
				ClientID:           "my-client",
			},
		),
	)
})
