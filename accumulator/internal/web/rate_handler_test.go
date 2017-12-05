package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/accumulator/internal/web"
	"github.com/gorilla/mux"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RateHandlers", func() {
	Describe("RateShow", func() {
		It("returns a 404 when timestamp is not a number", func() {
			h := web.RatesShow(&rateStore{})
			router := mux.NewRouter()
			router.Handle("/state/{timestamp}", h)

			r, err := http.NewRequest("", "/state/not-a-number", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("returns a 404 when a rate is not found", func() {
			h := web.RatesShow(&rateStore{rateError: errors.New("not found")})
			router := mux.NewRouter()
			router.Handle("/state/{timestamp}", h)

			r, err := http.NewRequest("", "/state/12345", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("returns a 400 when there is no timestamp string", func() {
			h := web.RatesShow(&rateStore{rateError: errors.New("not found")})

			r, err := http.NewRequest("", "", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
