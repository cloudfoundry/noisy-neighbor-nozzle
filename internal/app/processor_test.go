package app_test

import (
	"code.cloudfoundry.org/noisy-neighbor-nozzle/internal/app"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processor", func() {
	It("reads from the next func and increments", func() {
		next := func() *events.Envelope {
			return logMessage
		}

		incIDs := make(chan string, 10)
		inc := func(id string) {
			incIDs <- id
		}

		p := app.NewProcessor(next, inc)
		go p.Run()

		Eventually(incIDs).Should(Receive(Equal("app-id/0")))
	})

	It("ignores envelopes that are not logs", func() {
		next := func() *events.Envelope {
			return httpStartStop
		}

		incIDs := make(chan string, 10)
		inc := func(id string) {
			incIDs <- id
		}

		p := app.NewProcessor(next, inc)
		go p.Run()

		Consistently(incIDs).ShouldNot(Receive())
	})
})

var (
	logMessage = &events.Envelope{
		EventType: events.Envelope_LogMessage.Enum(),
		LogMessage: &events.LogMessage{
			AppId:          proto.String("app-id"),
			SourceInstance: proto.String("0"),
		},
	}

	httpStartStop = &events.Envelope{
		EventType: events.Envelope_HttpStartStop.Enum(),
	}
)
