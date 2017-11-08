package app_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/internal/app"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Noisyneighbor", func() {
	It("serves an HTTP endpoint with the noisiest applications", func() {
		uaa := newSpyUAA()
		defer uaa.stop()

		loggregator := newFakeLoggregator(testEnvelopes)
		cfg := app.Config{
			LoggregatorAddr: strings.Replace(loggregator.server.URL, "http", "ws", -1),
			BufferSize:      1000,
			PollingInterval: 100 * time.Millisecond,
			UAAAddr:         uaa.server.URL,
		}
		nn := app.New(cfg)
		Expect(uaa.tokenCalled()).To(Equal(int64(1)))

		go nn.Run()
		defer nn.Stop()

		Eventually(func() error {
			req, err := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://%s/state", nn.Addr()),
				nil,
			)

			_, err = http.DefaultClient.Do(req)

			return err
		}).Should(Succeed())
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
			Timestamp: proto.Int64(1234),
			Origin:    proto.String("origin"),
			EventType: events.Envelope_LogMessage.Enum(),
			LogMessage: &events.LogMessage{
				Timestamp:   proto.Int64(1234),
				AppId:       proto.String("app-id-1"),
				Message:     []byte(""),
				MessageType: events.LogMessage_OUT.Enum(),
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

type spyUAA struct {
	_tokenCalled int64
	server       *httptest.Server
}

func newSpyUAA() *spyUAA {
	s := &spyUAA{}
	s.server = httptest.NewServer(s)

	return s
}

func (s *spyUAA) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&s._tokenCalled, 1)
	w.Write([]byte(`{
		"access_token" : "68234a8da3b1436ba1b8450adde11e11",
		"token_type" : "bearer",
		"expires_in" : 43199,
		"scope" : "clients.read emails.write scim.userids password.write idps.write notifications.write oauth.login scim.write critical_notifications.write",
		"jti" : "68234a8da3b1436ba1b8450adde11e11"
	}`))
}

func (s *spyUAA) tokenCalled() int64 {
	return atomic.LoadInt64(&s._tokenCalled)
}

func (s *spyUAA) stop() {
	s.server.CloseClientConnections()
	s.server.Close()
}
