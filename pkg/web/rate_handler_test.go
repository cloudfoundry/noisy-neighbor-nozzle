package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/web"
	"github.com/gorilla/mux"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RateHandlers", func() {
	Describe("RatesShow", func() {
		It("returns a 404 when timestamp is not a number", func() {
			h := web.RatesShow(&rateStore{}, time.Minute)
			router := mux.NewRouter()
			router.Handle("/rates/{timestamp}", h)

			r, err := http.NewRequest("", "/rates/not-a-number", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("returns a 404 when a rate is not found", func() {
			h := web.RatesShow(&rateStore{rateError: errors.New("not found")}, time.Minute)
			router := mux.NewRouter()
			router.Handle("/rates/{timestamp}", h)

			r, err := http.NewRequest("", "/rates/12345", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("returns a 400 when there is no timestamp string", func() {
			h := web.RatesShow(&rateStore{rateError: errors.New("not found")}, time.Minute)

			r, err := http.NewRequest("", "", nil)
			Expect(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		Describe("query parameters", func() {
			It("truncates the timestamp", func() {
				rs := &rateStore{}
				h := web.RatesShow(rs, time.Minute)
				router := mux.NewRouter()
				router.Handle("/rates/{timestamp}", h)

				r, err := http.NewRequest(
					http.MethodGet,
					"/rates/1515426389?truncate_timestamp=true",
					nil,
				)
				Expect(err).ToNot(HaveOccurred())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(rs.rateTimestamp).To(Equal(int64(1515426360)))
			})
		})
	})
})
