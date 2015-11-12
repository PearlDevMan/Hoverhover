package main

import (
	"bytes"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"net/http"
)

// AllRecordsHandler returns JSON content type http response
func (d *DBClient) AllRecordsHandler(req *http.Request) *http.Response {

	records, err := d.getAllRecordsRaw()
func getBoneRouter(d DBClient) *bone.Mux {
	mux := bone.New()
	mux.Get("/records", http.HandlerFunc(d.AllRecordsHandler))
	// handling static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// concatenating string

	if err == nil {
		newResponse := &http.Response{}
		newResponse.Request = req

		newResponse.Header.Set("Content-Type", "application/json")

		// adding body
		var buff bytes.Buffer

		for _, record := range records {
			buff.WriteString(record)
		}

		buf := bytes.NewBuffer(buff.Bytes())
		newResponse.ContentLength = int64(buf.Len())
		newResponse.Body = ioutil.NopCloser(buf)

		newResponse.StatusCode = 200

		return newResponse

	} else {

		return goproxy.NewResponse(req,
			goproxy.ContentTypeText, http.StatusInternalServerError,
			"Failed to retrieve records from cache!")
	}

}
