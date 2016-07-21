package matching

import (
	"github.com/SpectoLabs/hoverfly/core/models"
	"errors"
	"net/http"
)

type RequestTemplateStore []RequestTemplatePayload


type RequestTemplatePayload struct {
	RequestTemplate RequestTemplate        `json: "requestTemplate"`
	Response        models.ResponseDetails `json: "response"`
}

type RequestTemplatePayloadJson struct {
	Data *[]RequestTemplatePayload `json:"data"`
}

type RequestTemplate struct {
	Path        string              `json:"path"`
	Method      string              `json:"method"`
	Destination string              `json:"destination"`
	Scheme      string              `json:"scheme"`
	Query       string              `json:"query"`
	Body        string              `json:"body"`
	Headers     map[string][]string `json:"headers"`
}

func(this *RequestTemplateStore) GetPayload(req *http.Request) (*models.Payload, error) {
	// iterate through the request templates, looking for template to match request
	for _, entry := range *this {
		// TODO: not matching by default on URL and body - need to enable this
		// TODO: need to enable regex matches
		if headerMatch(entry.RequestTemplate.Headers, req.Header) {
			// return the first template to match
			return &models.Payload{Response: entry.Response}, nil
		}
	}
	return nil, errors.New("No match found")
}

/**
Check keys and corresponding values in template headers are also present in request headers
 */
func headerMatch(tmplHeaders map[string][]string, reqHeaders http.Header) (bool) {

	for headerName, headerVal := range tmplHeaders {
		// TODO: case insensitive lookup
		// TODO: whole slice equality check
		reqHeaderVal, ok := reqHeaders[headerName]
		if (ok && headerVal[0] == reqHeaderVal[0]) {
			continue;
		} else {
			return false;
		}
	}
	return true;
}