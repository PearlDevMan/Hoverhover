package hoverfly_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/SpectoLabs/hoverfly/core/views"
	"github.com/SpectoLabs/hoverfly/functional-tests"
	"github.com/dghubble/sling"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("When I run Hoverfly", func() {

	var (
		hoverfly *functional_tests.Hoverfly
	)

	BeforeEach(func() {
		hoverfly = functional_tests.NewHoverfly()
		hoverfly.Start()
	})

	AfterEach(func() {
		hoverfly.Stop()
	})

	Context("When running in capture mode", func() {

		BeforeEach(func() {
			hoverfly.SetMode("capture")
		})

		Context("without middleware", func() {

			It("Should capture the request and response", func() {

				fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.Header().Set("Date", "date")
					w.Write([]byte("Hello world"))
				}))

				defer fakeServer.Close()

				resp := hoverfly.Proxy(sling.New().Get(fakeServer.URL))
				Expect(resp.StatusCode).To(Equal(200))

				expectedDestination := strings.Replace(fakeServer.URL, "http://", "", 1)

				recordsJson, err := ioutil.ReadAll(hoverfly.GetSimulation())
				Expect(err).To(BeNil())
				Expect(recordsJson).To(MatchJSON(fmt.Sprintf(
					`{
					  "data": [
					    {
					      "response": {
						"status": 200,
						"body": "Hello world",
						"encodedBody": false,
						"headers": {
						  "Content-Length": [
						    "11"
						  ],
						  "Content-Type": [
						    "text/plain"
						  ],
						  "Date": [
						    "date"
						  ],
						  "Hoverfly": [
						    "Was-Here"
						  ]
						}
					      },
					      "request": {
					      	"requestType": "recording",
						"path": "/",
						"method": "GET",
						"destination": "%v",
						"scheme": "http",
						"query": "",
						"body": "",
						"headers": {
						  "Accept-Encoding": [
						    "gzip"
						  ],
						  "User-Agent": [
						    "Go-http-client/1.1"
						  ]
						}
					      }
					    }
					  ]
					}`, expectedDestination)))
			})
		})

		It("Should capture a redirect response", func() {

			fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Date", "date")
				w.Write([]byte("redirection got you here"))
			}))
			fakeServerUrl, _ := url.Parse(fakeServer.URL)
			fakeServerRedirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", fakeServer.URL)
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(301)
			}))
			fakeServerRedirectUrl, _ := url.Parse(fakeServerRedirect.URL)

			defer fakeServer.Close()
			defer fakeServerRedirect.Close()

			resp := hoverfly.Proxy(sling.New().Get(fakeServerRedirect.URL))
			Expect(resp.StatusCode).To(Equal(301))

			expectedRedirectDestination := strings.Replace(fakeServerRedirectUrl.String(), "http://", "", 1)

			recordsJson, err := ioutil.ReadAll(hoverfly.GetSimulation())
			Expect(err).To(BeNil())

			payload := views.RequestResponsePairPayload{}

			json.Unmarshal(recordsJson, &payload)
			Expect(payload.Data).To(HaveLen(1))

			Expect(payload.Data[0].Request.Destination).To(Equal(expectedRedirectDestination))

			Expect(payload.Data[0].Response.Status).To(Equal(301))
			Expect(payload.Data[0].Response.Headers["Location"][0]).To(Equal(fakeServerUrl.String()))
		})
	})
})
