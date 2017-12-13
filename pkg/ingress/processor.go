package ingress

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

const (
	sourceTypeRouter = "RTR"
)

// Next is a func that reads an envelope off of a buffer.
type Next func() *events.Envelope

// Inc is a func that updates a counter for a given ID
type Inc func(string)

// Processor will read data from the Diode and update values in the store
type Processor struct {
	next              Next
	inc               Inc
	includeRouterLogs bool
}

// NewProcessor initializes a new Processor.
func NewProcessor(n Next, i Inc, includeRouterLogs bool) *Processor {
	return &Processor{
		next:              n,
		inc:               i,
		includeRouterLogs: includeRouterLogs,
	}
}

// Run will read events.Envelopes from the processors next func and increment
// the counter for the Envelopes source instance. This is a blocking method that
// will run indefinitely.
func (p *Processor) Run() {
	for {
		e := p.next()

		if e.GetEventType() != events.Envelope_LogMessage {
			continue
		}

		l := e.GetLogMessage()

		if !p.includeRouterLogs && l.GetSourceType() == sourceTypeRouter {
			continue
		}

		p.inc(fmt.Sprintf("%s/%s", l.GetAppId(), l.GetSourceInstance()))
	}
}
