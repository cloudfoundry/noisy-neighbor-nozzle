package app

import (
	"github.com/cloudfoundry/sonde-go/events"
)

// Next is a func that reads an envelope off of a buffer.
type Next func() *events.Envelope

// Inc is a func that updates a counter for a given ID
type Inc func(string)

// Processor will read data from the Diode and update values in the store
type Processor struct {
	next Next
	inc  Inc
}

// NewProcessor initializes a new Processor.
func NewProcessor(n Next, i Inc) *Processor {
	return &Processor{
		next: n,
		inc:  i,
	}
}

// Run will read events.Envelopes from the processors next func and increment
// the counter for the Envelopes source instance. This is a blocking method that
// will run indefinately.
func (p *Processor) Run() {
	for {
		e := p.next()

		if e.GetEventType() != events.Envelope_LogMessage {
			continue
		}

		p.inc(e.GetLogMessage().GetSourceInstance())
	}
}
