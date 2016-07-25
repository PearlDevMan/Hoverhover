package hoverfly

import (
	. "github.com/onsi/gomega"
	"github.com/SpectoLabs/hoverfly/core/models"
	"github.com/SpectoLabs/hoverfly/core/testutil"
	"testing"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http/httptest"
	"encoding/json"
)

func TestChangeBodyMiddleware(t *testing.T) {
	command := "./examples/middleware/modify_response/modify_response.py"

	resp := models.ResponseDetails{Status: 201, Body: "original body"}
	req := models.RequestDetails{Path: "/", Method: "GET", Destination: "hostname-x", Query: ""}

	payload := models.Payload{Response: resp, Request: req}

	newPayload, err := ExecuteMiddlewareLocally(command, payload)

	testutil.Expect(t, err, nil)
	testutil.Expect(t, newPayload.Response.Body, "body was replaced by middleware\n")
}

func TestMalformedPayloadMiddleware(t *testing.T) {
	command := "./examples/middleware/ruby_echo/echo.rb"

	resp := models.ResponseDetails{Status: 201, Body: "original body"}
	req := models.RequestDetails{Path: "/", Method: "GET", Destination: "hostname-x", Query: ""}

	payload := models.Payload{Response: resp, Request: req}

	newPayload, err := ExecuteMiddlewareLocally(command, payload)

	testutil.Expect(t, err, nil)
	testutil.Expect(t, newPayload.Response.Body, "original body")
}

func TestMakeCustom404(t *testing.T) {
	command := "go run ./examples/middleware/go_example/change_to_custom_404.go"

	resp := models.ResponseDetails{Status: 201, Body: "original body"}
	req := models.RequestDetails{Path: "/", Method: "GET", Destination: "hostname-x", Query: ""}

	payload := models.Payload{Response: resp, Request: req}

	newPayload, err := ExecuteMiddlewareLocally(command, payload)

	testutil.Expect(t, err, nil)
	testutil.Expect(t, newPayload.Response.Body, "Custom body here")
	testutil.Expect(t, newPayload.Response.Status, 404)
	testutil.Expect(t, newPayload.Response.Headers["middleware"][0], "changed response")
}

func TestReflectBody(t *testing.T) {
	command := "./examples/middleware/reflect_body/reflect_body.py"

	req := models.RequestDetails{Path: "/", Method: "GET", Destination: "hostname-x", Query: "", Body: "request_body_here"}

	payload := models.Payload{Request: req}

	newPayload, err := ExecuteMiddlewareLocally(command, payload)

	testutil.Expect(t, err, nil)
	testutil.Expect(t, newPayload.Response.Body, req.Body)
	testutil.Expect(t, newPayload.Request.Method, req.Method)
	testutil.Expect(t, newPayload.Request.Destination, req.Destination)
}

func processHandlerOkay(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	var newPayloadView models.PayloadView

	json.Unmarshal(body, &newPayloadView)

	newPayloadView.Response.Body = "You got straight up messed with"

	bts, _ := json.Marshal(newPayloadView)
	w.Write(bts)
}

func processHandlerOkayButNoResponse(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func processHandlerNotOkay(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

func TestExecuteMiddlewareRemotely(t *testing.T) {
	RegisterTestingT(t)

	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/process", processHandlerOkay).Methods("POST")
	server := httptest.NewServer(muxRouter)
	defer server.Close()

	testPayload := models.Payload{
		Response: models.ResponseDetails{
			Body: "Normal body",
		},
	}

	processedPayload, err := ExecuteMiddlewareRemotely(server.URL + "/process", testPayload)
	Expect(err).To(BeNil())

	Expect(processedPayload).ToNot(Equal(testPayload))
	Expect(processedPayload.Response.Body).To(Equal("You got straight up messed with"))
}

func TestExecuteMiddlewareRemotely_ReturnsErrorIfDoesntGetA200_AndSamePayload(t *testing.T) {
	RegisterTestingT(t)
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/process", processHandlerNotOkay).Methods("POST")
	server := httptest.NewServer(muxRouter)
	defer server.Close()

	testPayload := models.Payload{
		Response: models.ResponseDetails{
			Body: "Normal body",
		},
	}

	processedPayload, err := ExecuteMiddlewareRemotely(server.URL + "/process", testPayload)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("Error when communicating with remote middleware"))

	Expect(processedPayload).To(Equal(testPayload))
}

func TestExecuteMiddlewareRemotely_ReturnsErrorIfNoPayloadOnResponse_AnOriginalPayloadIsReturned(t *testing.T) {
	RegisterTestingT(t)
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/process", processHandlerOkayButNoResponse).Methods("POST")
	server := httptest.NewServer(muxRouter)
	defer server.Close()

	testPayload := models.Payload{
		Response: models.ResponseDetails{
			Body: "Normal body",
		},
	}

	processedPayload, err := ExecuteMiddlewareRemotely(server.URL + "/process", testPayload)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("unexpected end of JSON input"))

	Expect(processedPayload).To(Equal(testPayload))
}