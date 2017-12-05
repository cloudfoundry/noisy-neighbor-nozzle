package middleware_test

import (
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/internal/middleware"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AdminAuthorizer", func() {
	It("calls the next handler if the token is valid", func() {
		var givenToken, givenScope string
		checkToken := func(token, scope string) bool {
			givenToken = token
			givenScope = scope
			return true
		}
		var stubCalled int
		stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stubCalled++
		})
		recorder := httptest.NewRecorder()
		handler := middleware.AdminAuthMiddleware(checkToken)(stub)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer valid-token")

		handler.ServeHTTP(recorder, req)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(stubCalled).To(Equal(1))
		Expect(givenToken).To(Equal("valid-token"))
		Expect(givenScope).To(Equal("doppler.firehose"))
	})

	It("returns a 401 Unauthorized if check token fails", func() {
		checkToken := func(token, scope string) bool {
			return false
		}
		var stubCalled int
		stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stubCalled++
		})
		recorder := httptest.NewRecorder()
		handler := middleware.AdminAuthMiddleware(checkToken)(stub)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer invalid-token")

		handler.ServeHTTP(recorder, req)

		Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		Expect(stubCalled).To(Equal(0))
	})

	It("returns a 400 if the Authorization header is malformed", func() {
		checkToken := func(token, scope string) bool { return false }
		var stubCalled int
		stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stubCalled++
		})
		recorder := httptest.NewRecorder()
		handler := middleware.AdminAuthMiddleware(checkToken)(stub)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer")

		handler.ServeHTTP(recorder, req)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(stubCalled).To(Equal(0))
	})
})
