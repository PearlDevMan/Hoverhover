package v2

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
)

type HoverflyLogsStub struct{}

func (this HoverflyLogsStub) GetLogsView() LogsView {
	return LogsView{[]map[string]interface{}{
		map[string]interface{}{
			"msg": "test",
		},
	}}
}

func (this HoverflyLogsStub) GetFilteredLogsView(limit int) LogsView {
	return LogsView{[]map[string]interface{}{
		map[string]interface{}{
			"msg": "limited",
		},
	}}
}

func (this HoverflyLogsStub) GetLogs() string {
	return "a line of logs"
}

func (this HoverflyLogsStub) GetFilteredLogs(limit int) string {
	return "filtered logs by " + strconv.Itoa(limit)
}

func Test_LogsHandler_Get_ReturnsLogsView(t *testing.T) {
	RegisterTestingT(t)

	stubHoverfly := &HoverflyLogsStub{}
	unit := LogsHandler{Hoverfly: stubHoverfly}

	request, err := http.NewRequest("GET", "", nil)
	Expect(err).To(BeNil())

	response := makeRequestOnHandler(unit.Get, request)

	Expect(response.Code).To(Equal(http.StatusOK))

	logsView, err := unmarshalLogsView(response.Body)
	Expect(err).To(BeNil())

	Expect(logsView.Logs[0]["msg"]).To(Equal("test"))
}

func Test_LogsHandler_Get_ReturnsLogsInPlaintext(t *testing.T) {
	RegisterTestingT(t)

	stubHoverfly := &HoverflyLogsStub{}
	unit := LogsHandler{Hoverfly: stubHoverfly}

	request, err := http.NewRequest("GET", "", nil)
	Expect(err).To(BeNil())

	request.Header.Set("Content-Type", "text/plain")

	response := makeRequestOnHandler(unit.Get, request)
	Expect(response.Code).To(Equal(http.StatusOK))

	logs, _ := ioutil.ReadAll(response.Body)

	Expect(string(logs)).To(Equal("a line of logs"))
}

func unmarshalLogsView(buffer *bytes.Buffer) (LogsView, error) {
	body, err := ioutil.ReadAll(buffer)
	if err != nil {
		return LogsView{}, err
	}

	var logsView LogsView

	err = json.Unmarshal(body, &logsView)
	if err != nil {
		return LogsView{}, err
	}

	return logsView, nil
}
