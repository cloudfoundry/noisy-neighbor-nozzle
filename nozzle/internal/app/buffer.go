package app

import (
	"log"

	"code.cloudfoundry.org/go-diodes"
	"github.com/cloudfoundry/sonde-go/events"
)

// Buffer is a simple wrapper around the go-diode for the events.Envelope type.
type Buffer struct {
	d *diodes.Poller
}

// NewBuffer initializes and returns with given size.
func NewBuffer(size int) *Buffer {
	d := &Buffer{}

	d.d = diodes.NewPoller(diodes.NewOneToOne(size, d))

	return d
}

// Set adds an envelope to the buffer.
func (d *Buffer) Set(e *events.Envelope) {
	d.d.Set(diodes.GenericDataType(e))
}

// TryNext reads and returns the next envelope. This method will not block, if
// there is no envelope to read it will return nil for the envelope and a boolean
// of false.
func (d *Buffer) TryNext() (*events.Envelope, bool) {
	e, ok := d.d.TryNext()
	if !ok {
		return nil, ok
	}

	return (*events.Envelope)(e), true
}

// Next reads and returns the next envelope. This method will block until an
// envelope is available.
func (d *Buffer) Next() *events.Envelope {
	e := d.d.Next()

	return (*events.Envelope)(e)
}

// Alert is used by the internal diode. When envelopes are dropped we simply log
// a message noting how many envelopes were dropped.
func (d *Buffer) Alert(missed int) {
	log.Printf("dropped %d envelopes", missed)
}
