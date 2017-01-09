package modes

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/SpectoLabs/hoverfly/core/models"
	. "github.com/onsi/gomega"
)

func Test_ReconstructResponse_ReturnsAResponseWithCorrectStatus(t *testing.T) {
	RegisterTestingT(t)

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	pair := models.RequestResponsePair{
		Response: models.ResponseDetails{
			Status: 404,
		},
	}

	response := ReconstructResponse(req, pair)

	Expect(response.StatusCode).To(Equal(404))
}

func Test_ReconstructResponse_ReturnsAResponseWithBody(t *testing.T) {
	RegisterTestingT(t)

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	pair := models.RequestResponsePair{
		Response: models.ResponseDetails{
			Body: "test body",
		},
	}

	response := ReconstructResponse(req, pair)

	responseBody, err := ioutil.ReadAll(response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(Equal("test body"))
}

func Test_ReconstructResponse_AddsHeadersToResponse(t *testing.T) {
	RegisterTestingT(t)

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	pair := models.RequestResponsePair{}

	headers := make(map[string][]string)
	headers["Header"] = []string{"one"}

	pair.Response.Headers = headers

	response := ReconstructResponse(req, pair)

	Expect(response.Header.Get("Header")).To(Equal(headers["Header"][0]))
}

func Test_ReconstructResponse_CanReturnACompleteHttpResponseWithAllFieldsFilled(t *testing.T) {
	RegisterTestingT(t)

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	pair := models.RequestResponsePair{
		Response: models.ResponseDetails{
			Status: 201,
			Body:   "test body",
		},
	}

	headers := make(map[string][]string)
	headers["Header"] = []string{"header test"}
	headers["Other"] = []string{"header"}
	pair.Response.Headers = headers

	response := ReconstructResponse(req, pair)

	Expect(response.StatusCode).To(Equal(201))

	responseBody, err := ioutil.ReadAll(response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(Equal("test body"))

	Expect(response.Header.Get("Header")).To(Equal(headers["Header"][0]))
	Expect(response.Header.Get("Other")).To(Equal(headers["Other"][0]))
}
