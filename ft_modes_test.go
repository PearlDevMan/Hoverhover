package hoverfly_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	//"compress/gzip"
	"io/ioutil"
	"github.com/SpectoLabs/hoverfly"
	//"compress/gzip"
	"fmt"
	"strings"
	"net/url"
	"os"
	"github.com/SpectoLabs/hoverfly/models"
	"github.com/dghubble/sling"
)

var _ = Describe("Running Hoverfly in various modes", func() {

	Context("When running in capture mode", func() {

		var fakeServer * httptest.Server
		var fakeServerUrl * url.URL

		Context("without middleware", func() {

			BeforeEach(func() {
				requestCache.DeleteData()
				fakeServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.Header().Set("Date", "date")
					w.Write([]byte("Hello world"))
				}))

				defer fakeServer.Close()

				fakeServerUrl, _ = url.Parse(fakeServer.URL)
				SetHoverflyMode(hoverfly.CaptureMode)
				resp := CallFakeServerThroughProxy(fakeServer)
				Expect(resp.StatusCode).To(Equal(200))
			})

			It("Should capture the request and response", func() {
				expectedDestination := strings.Replace(fakeServerUrl.String(), "http://", "", 1)

				recordsJson, err := ioutil.ReadAll(ExportHoverflyRecords())
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

		Context("with middleware", func() {
			BeforeEach(func() {
				requestCache.DeleteData()
				fakeServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.Header().Set("Date", "date")
					w.Write([]byte("Hello world"))
				}))

				fakeServerUrl, _ = url.Parse(fakeServer.URL)
				SetHoverflyMode(hoverfly.CaptureMode)

				wd, err := os.Getwd()
				Expect(err).To(BeNil())
				hf.Cfg.Middleware = wd + "/testdata/middleware.py"
			})

			It("Should modify the request but not the response", func() {
				CallFakeServerThroughProxy(fakeServer)
				expectedDestination := strings.Replace(fakeServerUrl.String(), "http://", "", 1)
				recordsJson, err := ioutil.ReadAll(ExportHoverflyRecords())
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
						"path": "/",
						"method": "GET",
						"destination": "%v",
						"scheme": "http",
						"query": "",
						"body": "CHANGED",
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

			AfterEach(func() {
				hf.Cfg.Middleware = ""
				fakeServer.Close()
			})
		})
	})

	Context("When running in simulate mode", func() {

		BeforeEach(func(){
			SetHoverflyMode(hoverfly.SimulateMode)
			requestCache.DeleteData()
			pl1 := models.Payload{
				Request: models.RequestDetails{
					Path:"/path1",
					Method:"GET",
					Destination:"www.virtual.com",
					Scheme:"http",
					Query:"",
					Body:"",
					Headers:map[string][]string{"Header": []string{"value1"}},
				},
				Response: models.ResponseDetails{
					Status: 201,
					Body: "body1",
					Headers:map[string][]string{"Header": []string{"value1"}},
				},
			}
			encoded, _ := pl1.Encode()
			requestCache.Set([]byte(pl1.Id()), encoded)
			pl2 := models.Payload{
				Request: models.RequestDetails{
					Path:"/path2",
					Method:"GET",
					Destination:"www.virtual.com",
					Scheme:"http",
					Query:"",
					Body:"",
					Headers:map[string][]string{"Header": []string{"value2"}},
				},
				Response: models.ResponseDetails{
					Status: 202,
					Body: "body2",
					Headers:map[string][]string{"Header": []string{"value2"}},
				},
			}
			encoded, _ = pl2.Encode()
			requestCache.Set([]byte(pl2.Id()), encoded)
		})

		Context("without middleware", func() {
			It("should return the cached response", func() {
				resp := DoRequestThroughProxy(sling.New().Get("http://www.virtual.com/path1"))
				Expect(resp.StatusCode).To(Equal(201))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(string(body)).To(Equal("body1"))
				Expect(resp.Header).To(HaveKeyWithValue("Header", []string{"value1"}))
			})
		})

		Context("with middleware", func() {

			BeforeEach(func() {
				wd, err := os.Getwd()
				Expect(err).To(BeNil())
				hf.Cfg.Middleware = wd + "/testdata/middleware.py"
			})

			It("should apply middleware to the cached response", func() {
				resp := DoRequestThroughProxy(sling.New().Get("http://www.virtual.com/path2"))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(string(body)).To(Equal("CHANGED"))
			})

			AfterEach(func() {
				hf.Cfg.Middleware = ""
			})
		})
	})

	Context("When running in synthesise mode", func() {

		BeforeEach(func() {
			SetHoverflyMode(hoverfly.SynthesizeMode)
		})

		Context("With middleware", func() {

			BeforeEach(func() {
				wd, err := os.Getwd()
				Expect(err).To(BeNil())
				hf.Cfg.Middleware = wd + "/testdata/middleware_synthesise.py"
			})

			It("Should generate responses using middleware", func() {
				resp := DoRequestThroughProxy(sling.New().Get("http://www.virtual.com/path2"))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).To(BeNil())
				Expect(string(body)).To(Equal("GENERATED"))
			})

			AfterEach(func() {
				hf.Cfg.Middleware = ""
			})
		})

		Context("Without middleware", func() {
			It("Should fail to generate responses using middleware", func() {
				resp := DoRequestThroughProxy(sling.New().Get("http://www.virtual.com/path2"))
				Expect(resp.StatusCode).To(Equal(503))
			})
		})

	})
})
