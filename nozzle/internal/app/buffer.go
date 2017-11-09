package app

import (
	"log"

	"code.cloudfoundry.org/go-diodes"
	"github.com/cloudfoundry/sonde-go/events"
)

type Buffer struct {
	d *diodes.Poller
}

func NewBuffer(size int) *Buffer {
	d := &Buffer{}

	d.d = diodes.NewPoller(diodes.NewOneToOne(size, d))

	return d
}

func (d *Buffer) Set(e *events.Envelope) {
	d.d.Set(diodes.GenericDataType(e))
}

func (d *Buffer) TryNext() (*events.Envelope, bool) {
	e, ok := d.d.TryNext()
	if !ok {
		return nil, ok
	}

	return (*events.Envelope)(e), true
}

func (d *Buffer) Next() *events.Envelope {
	e := d.d.Next()

	return (*events.Envelope)(e)
}

func (d *Buffer) Alert(missed int) {
	log.Printf("dropped %d envelopes", missed)
}
