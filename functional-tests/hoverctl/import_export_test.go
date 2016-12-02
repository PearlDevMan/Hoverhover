package hoverctl_end_to_end

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"net/http"
	"net/http/httptest"

	"github.com/dghubble/sling"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
)

var _ = Describe("When I use hoverctl", func() {
	var (
		hoverflyCmd *exec.Cmd

		workingDir, _     = os.Getwd()
		adminPort         = freeport.GetPort()
		adminPortAsString = strconv.Itoa(adminPort)

		proxyPort = freeport.GetPort()

		v1HoverflyData = `
					{
						"data": [{
							"request": {
								"requestType": "recording",
								"path": "/api/bookings",
								"method": "POST",
								"destination": "www.my-test.com",
								"scheme": "http",
								"query": "",
								"body": "{\"flightId\": \"1\"}",
								"headers": {
									"Content-Type": [
										"application/json"
									]
								}
							},
							"response": {
								"status": 201,
								"body": "",
								"encodedBody": false,
								"headers": {
									"Location": [
										"http://localhost/api/bookings/1"
									]
								}
							}
						}]
					}`

		v2HoverflyData = `
			{
				"data": {
					"pairs": [{
						"response": {
							"status": 201,
							"body": "",
							"encodedBody": false,
							"headers": {
								"Location": ["http://localhost/api/bookings/1"]
							}
						},
						"request": {
							"requestType": "recording",
							"path": "/api/bookings",
							"method": "POST",
							"destination": "www.my-test.com",
							"scheme": "http",
							"query": "",
							"body": "{\"flightId\": \"1\"}",
							"headers": {
								"Content-Type": ["application/json"]
							}
						}
					}],
					"globalActions": {
						"delays": []
					}
				},
				"meta": {
					"schemaVersion": "v1",
					"hoverflyVersion": "v0.9.0",
					"timeExported": "2016-11-10T12:27:46Z"
				}
			}`

		v2HoverflySimulation = `"pairs":[{"response":{"status":201,"body":"","encodedBody":false,"headers":{"Location":["http://localhost/api/bookings/1"]}},"request":{"requestType":"recording","path":"/api/bookings","method":"POST","destination":"www.my-test.com","scheme":"http","query":"","body":"{\"flightId\": \"1\"}","headers":{"Content-Type":["application/json"]}}}],"globalActions":{"delays":[]}}`

		v2HoverflyMeta = `"meta":{"schemaVersion":"v1","hoverflyVersion":"v0.9.0","timeExported":`
	)

	Describe("with a running hoverfly", func() {

		BeforeEach(func() {
			hoverflyCmd = startHoverfly(adminPort, proxyPort, workingDir)
		})

		AfterEach(func() {
			hoverflyCmd.Process.Kill()
		})

		Describe("Managing Hoverflies data using the CLI", func() {

			BeforeEach(func() {
				DoRequest(sling.New().Post(fmt.Sprintf("http://localhost:%v/api/records", adminPort)).Body(strings.NewReader(v1HoverflyData)))

				resp := DoRequest(sling.New().Get(fmt.Sprintf("http://localhost:%v/api/records", adminPort)))
				bytes, _ := ioutil.ReadAll(resp.Body)
				Expect(string(bytes)).ToNot(Equal(`{"data":null}`))
			})

			It("can export", func() {

				fileName := generateFileName()
				// Export the data
				output, _ := exec.Command(hoverctlBinary, "export", fileName, "--admin-port="+adminPortAsString).Output()

				Expect(output).To(ContainSubstring("Successfully exported to " + fileName))

				data, err := ioutil.ReadFile(fileName)
				Expect(err).To(BeNil())

				Expect(string(data)).To(ContainSubstring(v2HoverflySimulation))
				Expect(string(data)).To(ContainSubstring(v2HoverflyMeta))
			})

			It("can import", func() {

				fileName := generateFileName()
				err := ioutil.WriteFile(fileName, []byte(v2HoverflyData), 0644)
				Expect(err).To(BeNil())

				output, _ := exec.Command(hoverctlBinary, "import", fileName, "--admin-port="+adminPortAsString).Output()

				Expect(output).To(ContainSubstring("Successfully imported from " + fileName))

				resp := DoRequest(sling.New().Get(fmt.Sprintf("http://localhost:%v/api/records", adminPort)))
				bytes, _ := ioutil.ReadAll(resp.Body)
				Expect(string(bytes)).To(MatchJSON(v1HoverflyData))
			})

			It("can import over http", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintln(w, v2HoverflyData)
				}))
				defer ts.Close()

				output, _ := exec.Command(hoverctlBinary, "import", ts.URL, "--admin-port="+adminPortAsString).Output()

				Expect(output).To(ContainSubstring("Successfully imported from " + ts.URL))

				resp := DoRequest(sling.New().Get(fmt.Sprintf("http://localhost:%v/api/records", adminPort)))
				bytes, _ := ioutil.ReadAll(resp.Body)
				Expect(string(bytes)).To(MatchJSON(v1HoverflyData))
			})

			It("can import v1 simulations", func() {

				fileName := generateFileName()
				err := ioutil.WriteFile(fileName, []byte(v1HoverflyData), 0644)
				Expect(err).To(BeNil())

				output, _ := exec.Command(hoverctlBinary, "import", "--v1", fileName, "--admin-port="+adminPortAsString).Output()

				Expect(output).To(ContainSubstring("Successfully imported from " + fileName))

				resp := DoRequest(sling.New().Get(fmt.Sprintf("http://localhost:%v/api/records", adminPort)))
				bytes, _ := ioutil.ReadAll(resp.Body)
				Expect(string(bytes)).To(MatchJSON(v1HoverflyData))
			})

			It("cannot import incorrect json / missing meta", func() {

				fileName := generateFileName()
				err := ioutil.WriteFile(fileName, []byte(`
				{
					"data": {
						"pairs": [{
							"response": {
								"status": 201,
								"body": "",
								"encodedBody": false,
								"headers": {
									"Location": ["http://localhost/api/bookings/1"]
								}
							},
							"request": {
								"requestType": "recording",
								"path": "/api/bookings",
								"method": "POST",
								"destination": "www.my-test.com",
								"scheme": "http",
								"query": "",
								"body": "{\"flightId\": \"1\"}",
								"headers": {
									"Content-Type": ["application/json"]
								}
							}
						}],
						"globalActions": {
							"delays": []
						}
					}
				}`), 0644)
				Expect(err).To(BeNil())

				output, _ := exec.Command(hoverctlBinary, "import", fileName, "--admin-port="+adminPortAsString).Output()

				Expect(output).To(ContainSubstring("Import to Hoverfly failed: Json did not match schema: Object->Key[meta].Value->Object"))

				resp := DoRequest(sling.New().Get(fmt.Sprintf("http://localhost:%v/api/records", adminPort)))
				bytes, _ := ioutil.ReadAll(resp.Body)
				Expect(string(bytes)).To(MatchJSON(v1HoverflyData))
			})
		})
	})
})
