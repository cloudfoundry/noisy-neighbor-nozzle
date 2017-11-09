package app

import "github.com/cloudfoundry/sonde-go/events"

// Set is a func used by the Ingestor to write envelopes to.
type Set func(*events.Envelope)

// Ingestor will read envelopes off of a channel and write them to a given buffer.
type Ingestor struct {
	msgs   <-chan *events.Envelope
	setter Set
}

// NewIngestor returns an initialized Ingestor.
func NewIngestor(msgs <-chan *events.Envelope, s Set) *Ingestor {
	return &Ingestor{
		msgs:   msgs,
		setter: s,
	}
}

// Run will start ingesting enveloeps off of the Ingestors message channel and
// writing them to the setter func. This method will block until the messages
// channel is closed.
func (i *Ingestor) Run() {
	for e := range i.msgs {
		i.setter(e)
	}
}
