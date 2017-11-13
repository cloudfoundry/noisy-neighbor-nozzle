package web_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/nozzle/internal/web"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateHandlers", func() {
	Describe("StateShow", func() {
		It("returns a 404 when timestamp is not a number", func() {
			h := web.StateShow(&rateStore{})

			r, err := http.NewRequest("", "", nil)
			Expect(err).ToNot(HaveOccurred())

			r = r.WithContext(
				context.WithValue(context.Background(), "timestamp", "not-a-number"),
			)

			Expect(r.Context().Value("timestamp")).To(Equal("not-a-number"))

			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("returns a 404 when a rate is not found", func() {
			h := web.StateShow(&rateStore{rateError: errors.New("not found")})

			r, err := http.NewRequest("", "", nil)
			Expect(err).ToNot(HaveOccurred())

			r = r.WithContext(
				context.WithValue(context.Background(), "timestamp", "12345"),
			)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("returns a 400 when there is no timestamp string", func() {
			h := web.StateShow(&rateStore{rateError: errors.New("not found")})

			r, err := http.NewRequest("", "", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
