package v2_test

import (
	"testing"

	"github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"github.com/SpectoLabs/hoverfly/core/util"
	. "github.com/onsi/gomega"
)

func Test_NewSimulationViewFromResponseBody_CanCreateSimulationFromV3Payload(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {
			"pairs": [
				{
					"response": {
						"status": 200,
						"body": "exact match",
						"encodedBody": false,
						"headers": {
							"Header": [
								"value"
							]
						},
						"templated":true
					},
					"request": {
						"destination": {
							"exactMatch": "test-server.com"
						}
					}
				}
			],
			"globalActions": {
				"delays": []
			}
		},
		"meta": {
			"schemaVersion": "v3",
			"hoverflyVersion": "v0.11.0",
			"timeExported": "2017-02-23T12:43:48Z"
		}
	}`))

	Expect(err).To(BeNil())

	Expect(simulation.RequestResponsePairs).To(HaveLen(1))

	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Body).To(BeNil())
	Expect(*simulation.RequestResponsePairs[0].RequestMatcher.Destination.ExactMatch).To(Equal("test-server.com"))
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Headers).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Method).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Path).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Query).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Scheme).To(BeNil())

	Expect(simulation.RequestResponsePairs[0].Response.Body).To(Equal("exact match"))
	Expect(simulation.RequestResponsePairs[0].Response.Templated).To(BeTrue())
	Expect(simulation.RequestResponsePairs[0].Response.EncodedBody).To(BeFalse())
	Expect(simulation.RequestResponsePairs[0].Response.Headers).To(HaveKeyWithValue("Header", []string{"value"}))
	Expect(simulation.RequestResponsePairs[0].Response.Status).To(Equal(200))

	Expect(simulation.SchemaVersion).To(Equal("v3"))
	Expect(simulation.HoverflyVersion).To(Equal("v0.11.0"))
	Expect(simulation.TimeExported).To(Equal("2017-02-23T12:43:48Z"))
}

func Test_NewSimulationViewFromResponseBody_CanCreateSimulationFromV2Payload(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {
			"pairs": [
				{
					"response": {
						"status": 200,
						"body": "exact match",
						"encodedBody": false,
						"headers": {
							"Header": [
								"value"
							]
						}
					},
					"request": {
						"destination": {
							"exactMatch": "test-server.com"
						}
					}
				}
			],
			"globalActions": {
				"delays": []
			}
		},
		"meta": {
			"schemaVersion": "v2",
			"hoverflyVersion": "v0.11.0",
			"timeExported": "2017-02-23T12:43:48Z"
		}
	}`))

	Expect(err).To(BeNil())

	Expect(simulation.RequestResponsePairs).To(HaveLen(1))

	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Body).To(BeNil())
	Expect(*simulation.RequestResponsePairs[0].RequestMatcher.Destination.ExactMatch).To(Equal("test-server.com"))
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Headers).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Method).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Path).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Query).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Scheme).To(BeNil())

	Expect(simulation.RequestResponsePairs[0].Response.Body).To(Equal("exact match"))
	Expect(simulation.RequestResponsePairs[0].Response.Templated).To(BeFalse())
	Expect(simulation.RequestResponsePairs[0].Response.EncodedBody).To(BeFalse())
	Expect(simulation.RequestResponsePairs[0].Response.Headers).To(HaveKeyWithValue("Header", []string{"value"}))
	Expect(simulation.RequestResponsePairs[0].Response.Status).To(Equal(200))

	Expect(simulation.SchemaVersion).To(Equal("v3"))
	Expect(simulation.HoverflyVersion).To(Equal("v0.11.0"))
	Expect(simulation.TimeExported).To(Equal("2017-02-23T12:43:48Z"))
}

func Test_NewSimulationViewFromResponseBody_WontCreateSimulationIfThereIsNoSchemaVersion(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {},
		"meta": {
			"hoverflyVersion": "v0.11.0",
			"timeExported": "2017-02-23T12:43:48Z"
		}
	}`))

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("Invalid JSON, missing \"meta.schemaVersion\" string"))

	Expect(simulation).ToNot(BeNil())
	Expect(simulation.RequestResponsePairs).To(HaveLen(0))
	Expect(simulation.GlobalActions.Delays).To(HaveLen(0))
}

func Test_NewSimulationViewFromResponseBody_WontBlowUpIfMetaIsMissing(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {}
	}`))

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal(`Invalid JSON, missing "meta" object`))

	Expect(simulation).ToNot(BeNil())
	Expect(simulation.RequestResponsePairs).To(HaveLen(0))
	Expect(simulation.GlobalActions.Delays).To(HaveLen(0))
}

func Test_NewSimulationViewFromResponseBody_CanCreateSimulationFromV1Payload(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {
			"pairs": [
				{
					"response": {
						"status": 200,
						"body": "exact match",
						"encodedBody": false,
						"headers": {
							"Header": [
								"value"
							]
						}
					},
					"request": {
						"destination":"test-server.com"
					}
				}
			],
			"globalActions": {
				"delays": []
			}
		},
		"meta": {
			"schemaVersion": "v1",
			"hoverflyVersion": "v0.11.0",
			"timeExported": "2017-02-23T12:43:48Z"
		}
	}`))

	Expect(err).To(BeNil())

	Expect(simulation.RequestResponsePairs).To(HaveLen(1))

	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Body).To(BeNil())
	Expect(*simulation.RequestResponsePairs[0].RequestMatcher.Destination.ExactMatch).To(Equal("test-server.com"))
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Headers).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Method).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Path).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Query).To(BeNil())
	Expect(simulation.RequestResponsePairs[0].RequestMatcher.Scheme).To(BeNil())

	Expect(simulation.RequestResponsePairs[0].Response.Body).To(Equal("exact match"))
	Expect(simulation.RequestResponsePairs[0].Response.EncodedBody).To(BeFalse())
	Expect(simulation.RequestResponsePairs[0].Response.Headers).To(HaveKeyWithValue("Header", []string{"value"}))
	Expect(simulation.RequestResponsePairs[0].Response.Status).To(Equal(200))
	Expect(simulation.RequestResponsePairs[0].Response.Templated).To(BeFalse())

	Expect(simulation.SchemaVersion).To(Equal("v3"))
	Expect(simulation.HoverflyVersion).To(Equal("v0.11.0"))
	Expect(simulation.TimeExported).To(Equal("2017-02-23T12:43:48Z"))
}

func Test_NewSimulationViewFromResponseBody_WontCreateSimulationFromInvalidV1Simulation(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {
			"pairs": [
				{
					
				}
			]
		},
		"meta": {
			"schemaVersion": "v1",
			"hoverflyVersion": "v0.11.0",
			"timeExported": "2017-02-23T12:43:48Z"
		}
	}`))

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("Invalid v1 simulation: request is required, response is required"))

	Expect(simulation).ToNot(BeNil())
	Expect(simulation.RequestResponsePairs).To(HaveLen(0))
	Expect(simulation.GlobalActions.Delays).To(HaveLen(0))
}

func Test_NewSimulationViewFromResponseBody_WontCreateSimulationFromUnknownSchemaVersion(t *testing.T) {
	RegisterTestingT(t)

	_, err := v2.NewSimulationViewFromResponseBody([]byte(`{
		"data": {
			"pairs": [
				{
					
				}
			]
		},
		"meta": {
			"schemaVersion": "r3",
			"hoverflyVersion": "v0.11.0",
			"timeExported": "2017-02-23T12:43:48Z"
		}
	}`))

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("Invalid simulation: schema version r3 is not supported by this version of Hoverfly, you may need to update Hoverfly"))
}

func Test_NewSimulationViewFromResponseBody_WontCreateSimulationFromInvalidJson(t *testing.T) {
	RegisterTestingT(t)

	simulation, err := v2.NewSimulationViewFromResponseBody([]byte(`{}{}[^.^]{}{}`))

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("Invalid JSON"))

	Expect(simulation).ToNot(BeNil())
	Expect(simulation.RequestResponsePairs).To(HaveLen(0))
	Expect(simulation.GlobalActions.Delays).To(HaveLen(0))
}

func Test_RequestDetailsView_GetQuery_SortsQueryString(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.RequestDetailsView{
		Query: util.StringToPointer("b=b&a=a"),
	}
	queryString := unit.GetQuery()
	Expect(queryString).ToNot(BeNil())

	Expect(*queryString).To(Equal("a=a&b=b"))
}

func Test_RequestDetailsView_GetQuery_ReturnsNilIfNil(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.RequestDetailsView{
		Query: nil,
	}
	queryString := unit.GetQuery()
	Expect(queryString).To(BeNil())
}

func Test_SimulationViewV1_Upgrade_ReturnsAV3Simulation(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.SimulationViewV1{
		v2.DataViewV1{
			RequestResponsePairViewV1: []v2.RequestResponsePairViewV1{
				{
					Request: v2.RequestDetailsView{
						RequestType: util.StringToPointer("recording"),
						Scheme:      util.StringToPointer("http"),
						Body:        util.StringToPointer("body"),
						Destination: util.StringToPointer("destination"),
						Method:      util.StringToPointer("GET"),
						Path:        util.StringToPointer("/path"),
						Query:       util.StringToPointer("query=query"),
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
					Response: v2.ResponseDetailsView{
						Status:      200,
						Body:        "body",
						EncodedBody: false,
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
				},
			},
		},
		v2.MetaView{
			SchemaVersion:   "v1",
			HoverflyVersion: "test",
			TimeExported:    "today",
		},
	}

	simulationViewV2 := unit.Upgrade()

	Expect(simulationViewV2.RequestResponsePairs).To(HaveLen(1))

	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Scheme).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("http"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Body).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("body"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Destination).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("destination"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Method).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("GET"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Path).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("/path"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Query).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("query=query"),
	}))
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Headers).To(BeEmpty())

	Expect(simulationViewV2.RequestResponsePairs[0].Response.Status).To(Equal(200))
	Expect(simulationViewV2.RequestResponsePairs[0].Response.Templated).To(BeFalse())
	Expect(simulationViewV2.RequestResponsePairs[0].Response.Body).To(Equal("body"))
	Expect(simulationViewV2.RequestResponsePairs[0].Response.EncodedBody).To(BeFalse())
	Expect(simulationViewV2.RequestResponsePairs[0].Response.Headers).To(HaveKeyWithValue("Test", []string{"headers"}))

	Expect(simulationViewV2.SchemaVersion).To(Equal("v3"))
	Expect(simulationViewV2.HoverflyVersion).To(Equal("test"))
	Expect(simulationViewV2.TimeExported).To(Equal("today"))
}

func Test_SimulationViewV1_Upgrade_ReturnsGlobMatchesIfTemplate(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.SimulationViewV1{
		v2.DataViewV1{
			RequestResponsePairViewV1: []v2.RequestResponsePairViewV1{
				{
					Request: v2.RequestDetailsView{
						RequestType: util.StringToPointer("template"),
						Scheme:      util.StringToPointer("http"),
						Body:        util.StringToPointer("body"),
						Destination: util.StringToPointer("destination"),
						Method:      util.StringToPointer("GET"),
						Path:        util.StringToPointer("/path"),
						Query:       util.StringToPointer("query=query"),
					},
					Response: v2.ResponseDetailsView{
						Status:      200,
						Body:        "body",
						EncodedBody: false,
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
				},
			},
		},
		v2.MetaView{
			SchemaVersion:   "v1",
			HoverflyVersion: "test",
			TimeExported:    "today",
		},
	}

	simulationViewV2 := unit.Upgrade()

	Expect(simulationViewV2.RequestResponsePairs).To(HaveLen(1))

	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Scheme).To(Equal(v2.RequestFieldMatchersView{
		GlobMatch: util.StringToPointer("http"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Body).To(Equal(v2.RequestFieldMatchersView{
		GlobMatch: util.StringToPointer("body"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Destination).To(Equal(v2.RequestFieldMatchersView{
		GlobMatch: util.StringToPointer("destination"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Method).To(Equal(v2.RequestFieldMatchersView{
		GlobMatch: util.StringToPointer("GET"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Path).To(Equal(v2.RequestFieldMatchersView{
		GlobMatch: util.StringToPointer("/path"),
	}))
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Query).To(Equal(v2.RequestFieldMatchersView{
		GlobMatch: util.StringToPointer("query=query"),
	}))
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Headers).To(BeEmpty())
}

func Test_SimulationViewV1_Upgrade_CanReturnAnIncompleteRequest(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.SimulationViewV1{
		v2.DataViewV1{
			RequestResponsePairViewV1: []v2.RequestResponsePairViewV1{
				{
					Request: v2.RequestDetailsView{
						Method: util.StringToPointer("POST"),
					},
					Response: v2.ResponseDetailsView{
						Status:      200,
						Body:        "body",
						EncodedBody: false,
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
				},
			},
		},
		v2.MetaView{
			SchemaVersion:   "v1",
			HoverflyVersion: "test",
			TimeExported:    "today",
		},
	}

	simulationViewV2 := unit.Upgrade()

	Expect(simulationViewV2.RequestResponsePairs).To(HaveLen(1))

	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Scheme).To(BeNil())
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Body).To(BeNil())
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Destination).To(BeNil())
	Expect(*simulationViewV2.RequestResponsePairs[0].RequestMatcher.Method).To(Equal(v2.RequestFieldMatchersView{
		ExactMatch: util.StringToPointer("POST"),
	}))
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Path).To(BeNil())
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Query).To(BeNil())
	Expect(simulationViewV2.RequestResponsePairs[0].RequestMatcher.Headers).To(BeNil())

	Expect(simulationViewV2.RequestResponsePairs[0].Response.Status).To(Equal(200))
	Expect(simulationViewV2.RequestResponsePairs[0].Response.Body).To(Equal("body"))
	Expect(simulationViewV2.RequestResponsePairs[0].Response.EncodedBody).To(BeFalse())
	Expect(simulationViewV2.RequestResponsePairs[0].Response.Headers).To(HaveKeyWithValue("Test", []string{"headers"}))
}

func Test_SimulationViewV1_Upgrade_UnescapesRequestQueryParameters(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.SimulationViewV1{
		v2.DataViewV1{
			RequestResponsePairViewV1: []v2.RequestResponsePairViewV1{
				{
					Request: v2.RequestDetailsView{
						Query: util.StringToPointer("q=10%20Downing%20Street%20London"),
					},
					Response: v2.ResponseDetailsView{
						Status:      200,
						Body:        "body",
						EncodedBody: false,
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
				},
			},
		},
		v2.MetaView{
			SchemaVersion:   "v1",
			HoverflyVersion: "test",
			TimeExported:    "today",
		},
	}

	simulationViewV3 := unit.Upgrade()

	Expect(simulationViewV3.RequestResponsePairs).To(HaveLen(1))
	Expect(*simulationViewV3.RequestResponsePairs[0].RequestMatcher.Query.ExactMatch).To(Equal("q=10 Downing Street London"))
}

func Test_SimulationViewV2_Upgrade_UnescapesExactMatchRequestQueryParameters(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.SimulationViewV2{
		v2.DataViewV2{
			RequestResponsePairs: []v2.RequestMatcherResponsePairViewV2{
				{
					RequestMatcher: v2.RequestMatcherViewV2{
						Query: &v2.RequestFieldMatchersView{
							ExactMatch: util.StringToPointer("q=10%20Downing%20Street%20London"),
						},
					},
					Response: v2.ResponseDetailsView{
						Status:      200,
						Body:        "body",
						EncodedBody: false,
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
				},
			},
		},
		v2.MetaView{
			SchemaVersion:   "v1",
			HoverflyVersion: "test",
			TimeExported:    "today",
		},
	}

	simulationViewV3 := unit.Upgrade()

	Expect(simulationViewV3.RequestResponsePairs).To(HaveLen(1))
	Expect(*simulationViewV3.RequestResponsePairs[0].RequestMatcher.Query.ExactMatch).To(Equal("q=10 Downing Street London"))
}

func Test_SimulationViewV2_Upgrade_UnescapesGlobMatchRequestQueryParameters(t *testing.T) {
	RegisterTestingT(t)

	unit := v2.SimulationViewV2{
		v2.DataViewV2{
			RequestResponsePairs: []v2.RequestMatcherResponsePairViewV2{
				{
					RequestMatcher: v2.RequestMatcherViewV2{
						Query: &v2.RequestFieldMatchersView{
							GlobMatch: util.StringToPointer("q=*%20London"),
						},
					},
					Response: v2.ResponseDetailsView{
						Status:      200,
						Body:        "body",
						EncodedBody: false,
						Headers: map[string][]string{
							"Test": []string{"headers"},
						},
					},
				},
			},
		},
		v2.MetaView{
			SchemaVersion:   "v2",
			HoverflyVersion: "test",
			TimeExported:    "today",
		},
	}

	simulationViewV3 := unit.Upgrade()

	Expect(simulationViewV3.RequestResponsePairs).To(HaveLen(1))
	Expect(*simulationViewV3.RequestResponsePairs[0].RequestMatcher.Query.GlobMatch).To(Equal("q=* London"))
}
