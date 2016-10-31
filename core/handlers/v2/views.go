package v2

import (
	"github.com/SpectoLabs/hoverfly/core/handlers/v1"
	"github.com/SpectoLabs/hoverfly/core/metrics"
	"github.com/SpectoLabs/hoverfly/core/interfaces"
)

type DestinationView struct {
	Destination string `json:"destination"`
}

type UsageView struct {
	Usage metrics.Stats `json:"usage"`
}

type MiddlewareView struct {
	Middleware string `json:"middleware"`
}

type ModeView struct {
	Mode string `json:"mode"`
}

type HoverflyView struct {
	DestinationView
	MiddlewareView
	ModeView
	UsageView
}

type SimulationView struct {
	DataView `json:"data"`
	MetaView `json:"meta"`
}

type DataView struct {
	RequestResponsePairs []RequestResponsePairView `json:"pairs"`
	GlobalActions        GlobalActionsView         `json:"globalActions"`
}

type RequestResponsePairView struct {
	Response ResponseDetailsView `json:"response"`
	Request  RequestDetailsView  `json:"request"`
}


//Gets Response - required for interfaces.RequestResponsePairView
func (this RequestResponsePairView) GetResponse() interfaces.Response { return this.Response }

//Gets Request - required for interfaces.RequestResponsePairView
func (this RequestResponsePairView) GetRequest() interfaces.Request { return this.Request }

// RequestDetailsView is used when marshalling and unmarshalling RequestDetails
type RequestDetailsView struct {
	RequestType *string             `json:"requestType"`
	Path        *string             `json:"path"`
	Method      *string             `json:"method"`
	Destination *string             `json:"destination"`
	Scheme      *string             `json:"scheme"`
	Query       *string             `json:"query"`
	Body        *string             `json:"body"`
	Headers     map[string][]string `json:"headers"`
}

//Gets RequestType - required for interfaces.Request
func (this RequestDetailsView) GetRequestType() *string { return this.RequestType }

//Gets Path - required for interfaces.Request
func (this RequestDetailsView) GetPath() *string { return this.Path }

//Gets Method - required for interfaces.Request
func (this RequestDetailsView) GetMethod() *string { return this.Method }

//Gets Destination - required for interfaces.Request
func (this RequestDetailsView) GetDestination() *string { return this.Destination }

//Gets Scheme - required for interfaces.Request
func (this RequestDetailsView) GetScheme() *string { return this.Scheme }

//Gets Query - required for interfaces.Request
func (this RequestDetailsView) GetQuery() *string { return this.Query }

//Gets Body - required for interfaces.Request
func (this RequestDetailsView) GetBody() *string { return this.Body }

//Gets Headers - required for interfaces.Request
func (this RequestDetailsView) GetHeaders() map[string][]string { return this.Headers }

// ResponseDetailsView is used when marshalling and
// unmarshalling requests. This struct's Body may be Base64
// encoded based on the EncodedBody field.
type ResponseDetailsView struct {
	Status      int                 `json:"status"`
	Body        string              `json:"body"`
	EncodedBody bool                `json:"encodedBody"`
	Headers     map[string][]string `json:"headers"`
}


//Gets Status - required for interfaces.Response
func (this ResponseDetailsView) GetStatus() int { return this.Status }

// Gets Body - required for interfaces.Response
func (this ResponseDetailsView) GetBody() string { return this.Body }

// Gets EncodedBody - required for interfaces.Response
func (this ResponseDetailsView) GetEncodedBody() bool { return this.EncodedBody }

// Gets Headers - required for interfaces.Response
func (this ResponseDetailsView) GetHeaders() map[string][]string { return this.Headers }

type GlobalActionsView struct {
	Delays []v1.ResponseDelayView `json:"delays"`
}

type MetaView struct {
	SchemaVersion   string `json:"schemaVersion"`
	HoverflyVersion string `json:"hoverflyVersion"`
	TimeExported    string `json:"timeExported"`
}
