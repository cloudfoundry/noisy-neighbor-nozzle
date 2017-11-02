package noisyneighbor_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/noisyneighbor"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Noisyneighbor", func() {
	It("serves an HTTP endpoint with the noisiest applications", func() {
		loggregator := newFakeLoggregator(testEnvelopes)
		cfg := noisyneighbor.Config{
			LoggregatorAddr: strings.Replace(loggregator.server.URL, "http", "ws", -1),
			BufferSize:      1000,
			BasicAuthCreds: noisyneighbor.BasicAuthCreds{
				Username: "username",
				Password: "password",
			},
		}
		nn := noisyneighbor.New(cfg)

		go nn.Run()
		defer nn.Stop()

		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("http://%s/offenders", nn.Addr()),
			nil,
		)
		req.SetBasicAuth(
			cfg.BasicAuthCreds.Username,
			cfg.BasicAuthCreds.Password,
		)

		resp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		Expect(body).To(MatchJSON(`[
			{"id": "app-id-1", "count": 3},
			{"id": "app-id-3", "count": 2},
			{"id": "app-id-2", "count": 1}
		]`))
	})
})

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	testEnvelopes = []*events.Envelope{
		{
			LogMessage: &events.LogMessage{
				AppId: proto.String("app-id-1"),
			},
		},
		{
			LogMessage: &events.LogMessage{
				AppId: proto.String("app-id-2"),
			},
		},
		{
			LogMessage: &events.LogMessage{
				AppId: proto.String("app-id-1"),
			},
		},
		{
			LogMessage: &events.LogMessage{
				AppId: proto.String("app-id-3"),
			},
		},
		{
			LogMessage: &events.LogMessage{
				AppId: proto.String("app-id-1"),
			},
		},
		{
			LogMessage: &events.LogMessage{
				AppId: proto.String("app-id-3"),
			},
		},
	}
)

type fakeLoggregator struct {
	envelopes []*events.Envelope
	close     chan struct{}
	server    *httptest.Server
}

func newFakeLoggregator(e []*events.Envelope) *fakeLoggregator {
	f := &fakeLoggregator{
		envelopes: e,
		close:     make(chan struct{}),
	}

	f.server = httptest.NewServer(f)

	return f
}

func (f *fakeLoggregator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	Expect(err).ToNot(HaveOccurred())

	for _, e := range f.envelopes {
		data, err := proto.Marshal(e)
		Expect(err).ToNot(HaveOccurred())

		err = conn.WriteMessage(websocket.BinaryMessage, data)
		Expect(err).ToNot(HaveOccurred())
	}

	<-f.close
}

func (f *fakeLoggregator) stop() {
	close(f.close)
}
