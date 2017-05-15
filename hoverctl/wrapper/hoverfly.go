package wrapper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/SpectoLabs/hoverfly/core/handlers"
	"github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"github.com/SpectoLabs/hoverfly/core/util"
	"github.com/kardianos/osext"
)

const (
	v1ApiDelays     = "/api/delays"
	v1ApiSimulation = "/api/records"

	v2ApiSimulation  = "/api/v2/simulation"
	v2ApiMode        = "/api/v2/hoverfly/mode"
	v2ApiDestination = "/api/v2/hoverfly/destination"
	v2ApiMiddleware  = "/api/v2/hoverfly/middleware"
	v2ApiCache       = "/api/v2/cache"
	v2ApiLogs        = "/api/v2/logs"
)

type APIStateSchema struct {
	Mode        string `json:"mode"`
	Destination string `json:"destination"`
}

type APIDelaySchema struct {
	Data []ResponseDelaySchema `json:"data"`
}

type ResponseDelaySchema struct {
	UrlPattern string `json:"urlpattern"`
	Delay      int    `json:"delay"`
	HttpMethod string `json:"httpmethod"`
}

type HoverflyAuthSchema struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HoverflyAuthTokenSchema struct {
	Token string `json:"token"`
}

type MiddlewareSchema struct {
	Middleware string `json:"middleware"`
}

type ErrorSchema struct {
	ErrorMessage string `json:"error"`
}

// Wipe will call the records endpoint in Hoverfly with a DELETE request, triggering Hoverfly to wipe the database
func DeleteSimulations(target Target) error {
	response, err := doRequest(target, "DELETE", v2ApiSimulation, "", nil)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Simulations were not deleted from Hoverfly")
	}

	return nil
}

// GetMode will go the state endpoint in Hoverfly, parse the JSON response and return the mode of Hoverfly
func GetMode(target Target) (string, error) {
	response, err := doRequest(target, "GET", v2ApiMode, "", nil)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	apiResponse := createAPIStateResponse(response)

	return apiResponse.Mode, nil
}

// Set will go the state endpoint in Hoverfly, sending JSON that will set the mode of Hoverfly
func SetModeWithArguments(target Target, modeView v2.ModeView) (string, error) {
	if modeView.Mode != "simulate" && modeView.Mode != "capture" &&
		modeView.Mode != "modify" && modeView.Mode != "synthesize" {
		return "", errors.New(modeView.Mode + " is not a valid mode")
	}
	bytes, err := json.Marshal(modeView)
	if err != nil {
		return "", err
	}

	response, err := doRequest(target, "PUT", v2ApiMode, string(bytes), nil)
	if err != nil {
		return "", err
	}

	if response.StatusCode == http.StatusBadRequest {
		return "", handlerError(response)
	}

	apiResponse := createAPIStateResponse(response)

	return apiResponse.Mode, nil
}

// GetDestination will go the destination endpoint in Hoverfly, parse the JSON response and return the destination of Hoverfly
func GetDestination(target Target) (string, error) {
	response, err := doRequest(target, "GET", v2ApiDestination, "", nil)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	apiResponse := createAPIStateResponse(response)

	return apiResponse.Destination, nil
}

// SetDestination will go the destination endpoint in Hoverfly, sending JSON that will set the destination of Hoverfly
func SetDestination(target Target, destination string) (string, error) {

	response, err := doRequest(target, "PUT", v2ApiDestination, `{"destination":"`+destination+`"}`, nil)
	if err != nil {
		return "", err
	}

	apiResponse := createAPIStateResponse(response)

	return apiResponse.Destination, nil
}

// GetMiddle will go the middleware endpoint in Hoverfly, parse the JSON response and return the middleware of Hoverfly
func GetMiddleware(target Target) (v2.MiddlewareView, error) {
	response, err := doRequest(target, "GET", v2ApiMiddleware, "", nil)
	if err != nil {
		return v2.MiddlewareView{}, err
	}

	defer response.Body.Close()

	middlewareResponse := createMiddlewareSchema(response)

	return middlewareResponse, nil
}

func SetMiddleware(target Target, binary, script, remote string) (v2.MiddlewareView, error) {
	middlewareRequest := &v2.MiddlewareView{
		Binary: binary,
		Script: script,
		Remote: remote,
	}

	marshalledMiddleware, err := json.Marshal(middlewareRequest)
	if err != nil {
		return v2.MiddlewareView{}, err
	}

	response, err := doRequest(target, "PUT", v2ApiMiddleware, string(marshalledMiddleware), nil)
	if err != nil {
		return v2.MiddlewareView{}, err
	}

	err = handleResponseError(response, "Hoverfly could not execute this middleware")
	if err != nil {
		return v2.MiddlewareView{}, err
	}

	apiResponse := createMiddlewareSchema(response)

	return apiResponse, nil
}

func GetLogs(target Target, format string) ([]string, error) {
	headers := map[string]string{
		"Content-Type": "text/plain",
	}

	if format == "json" {
		headers["Content-Type"] = "application/json"
	}

	response, err := doRequest(target, "GET", v2ApiLogs, "", headers)
	if err != nil {
		return []string{}, err
	}

	defer response.Body.Close()

	responseBody, _ := ioutil.ReadAll(response.Body)
	if format == "json" {
		trimmedBody := responseBody[9 : len(responseBody)-2]
		return strings.SplitAfter(string(trimmedBody), "},"), nil
	} else {
		return strings.Split(string(responseBody), "\n"), nil
	}
}

func ImportSimulation(target Target, simulationData string) error {
	response, err := doRequest(target, "PUT", v2ApiSimulation, simulationData, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		var errorView ErrorSchema
		json.Unmarshal(body, &errorView)
		return errors.New("Import to Hoverfly failed: " + errorView.ErrorMessage)
	}

	return nil
}

func FlushCache(target Target) error {
	response, err := doRequest(target, "DELETE", v2ApiCache, "", nil)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New("Cache was not set on Hoverfly")
	}

	return nil
}

func ExportSimulation(target Target) ([]byte, error) {
	response, err := doRequest(target, "GET", v2ApiSimulation, "", nil)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Debug(err.Error())
		return nil, errors.New("Could not export from Hoverfly")
	}

	var jsonBytes bytes.Buffer
	err = json.Indent(&jsonBytes, body, "", "\t")
	if err != nil {
		log.Debug(err.Error())
		return nil, errors.New("Could not export from Hoverfly")
	}

	return jsonBytes.Bytes(), nil
}

func createAPIStateResponse(response *http.Response) APIStateSchema {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Debug(err.Error())
	}

	var apiResponse APIStateSchema

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Debug(err.Error())
	}

	return apiResponse
}

func createMiddlewareSchema(response *http.Response) v2.MiddlewareView {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Debug(err.Error())
	}

	var middleware v2.MiddlewareView

	err = json.Unmarshal(body, &middleware)
	if err != nil {
		log.Debug(err.Error())
	}

	return middleware
}

func Login(target Target, username, password string) (string, error) {
	credentials := HoverflyAuthSchema{
		Username: username,
		Password: password,
	}

	jsonCredentials, err := json.Marshal(credentials)
	if err != nil {
		return "", fmt.Errorf("There was an error when preparing to login")
	}

	request, err := http.NewRequest("POST", BuildURL(target, "/api/token-auth"), strings.NewReader(string(jsonCredentials)))
	if err != nil {
		return "", fmt.Errorf("There was an error when preparing to login")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("There was an error when logging in")
	}

	if response.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("Too many failed login attempts, please wait 10 minutes")
	}

	if response.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("Incorrect username or password")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("There was an error when logging in")
	}

	var authToken HoverflyAuthTokenSchema
	err = json.Unmarshal(body, &authToken)
	if err != nil {
		return "", fmt.Errorf("There was an error when logging in")
	}

	return authToken.Token, nil
}

func BuildURL(target Target, endpoint string) string {
	if !strings.HasPrefix(target.Host, "http://") && !strings.HasPrefix(target.Host, "https://") {
		if IsLocal(target.Host) {
			return fmt.Sprintf("http://%v:%v%v", target.Host, target.AdminPort, endpoint)
		} else {
			return fmt.Sprintf("https://%v:%v%v", target.Host, target.AdminPort, endpoint)
		}
	}
	return fmt.Sprintf("%v:%v%v", target.Host, target.AdminPort, endpoint)
}

func IsLocal(url string) bool {
	return strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1")
}

/*
This isn't working as intended, its working, just not how I imagined it.
*/

func runBinary(target *Target, path string, hoverflyDirectory HoverflyDirectory) (*exec.Cmd, error) {
	flags := target.BuildFlags()

	cmd := exec.Command(path, flags...)
	log.Debug(cmd.Args)
	file, err := os.Create(hoverflyDirectory.Path + "/hoverfly." + strconv.Itoa(target.AdminPort) + "." + strconv.Itoa(target.ProxyPort) + ".log")
	if err != nil {
		log.Debug(err)
		return nil, errors.New("Could not create log file")
	}

	cmd.Stdout = file
	cmd.Stderr = file
	defer file.Close()

	err = cmd.Start()
	if err != nil {
		log.Debug(err)
		return nil, errors.New("Could not start Hoverfly")
	}

	return cmd, nil
}

func Start(target *Target, hoverflyDirectory HoverflyDirectory) error {
	err := checkPorts(target.AdminPort, target.ProxyPort)
	if err != nil {
		return err
	}

	binaryLocation, err := osext.ExecutableFolder()
	if err != nil {
		log.Debug(err)
		return errors.New("Could not start Hoverfly")
	}

	cmd, err := runBinary(target, binaryLocation+"/hoverfly", hoverflyDirectory)
	if err != nil {
		cmd, err = runBinary(target, "hoverfly", hoverflyDirectory)
		if err != nil {
			return errors.New("Could not start Hoverfly")
		}
	}

	timeout := time.After(10 * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	statusCode := 0

	for {
		select {
		case <-timeout:
			if err != nil {
				log.Debug(err)
			}
			return errors.New(fmt.Sprintf("Timed out waiting for Hoverfly to become healthy, returns status: %v", statusCode))
		case <-tick:
			resp, err := http.Get(fmt.Sprintf("http://localhost:%v/api/health", target.AdminPort))
			if err == nil {
				statusCode = resp.StatusCode
			} else {
				statusCode = 0
			}
		}

		if statusCode == 200 {
			break
		}
	}

	target.Pid = cmd.Process.Pid

	return nil
}

func Stop(target *Target, hoverflyDirectory HoverflyDirectory) error {
	hoverflyProcess := os.Process{Pid: target.Pid}
	err := hoverflyProcess.Kill()
	if err != nil {
		log.Info(err.Error())
		return errors.New("Could not kill Hoverfly [process " + strconv.Itoa(target.Pid) + "]")
	}

	target.Pid = 0

	return nil
}

func doRequest(target Target, method, url, body string, headers map[string]string) (*http.Response, error) {
	url = BuildURL(target, url)

	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Could not connect to Hoverfly at %v:%v", target.Host, target.AdminPort)
	}

	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}

	if target.AuthToken != "" {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %v", target.AuthToken))
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to Hoverfly at %v:%v", target.Host, target.AdminPort)
	}

	if response.StatusCode == 401 {
		return nil, errors.New("Hoverfly requires authentication\n\nRun `hoverctl login -t " + target.Name + "`")
	}

	return response, nil
}

func checkPorts(ports ...int) error {
	for _, port := range ports {
		server, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return fmt.Errorf("Could not start Hoverfly\n\nPort %v was not free", port)
		}
		server.Close()
	}

	return nil
}

func handlerError(response *http.Response) error {
	responseBody, err := util.GetResponseBody(response)
	if err != nil {
		return errors.New("Error when communicating with Hoverfly")
	}

	var errorView handlers.ErrorView
	err = json.Unmarshal([]byte(responseBody), &errorView)
	if err != nil {
		return errors.New("Error when communicating with Hoverfly")
	}

	return errors.New(errorView.Error)
}

func handleResponseError(response *http.Response, errorMessage string) error {
	if response.StatusCode != 200 {
		defer response.Body.Close()
		responseError, _ := ioutil.ReadAll(response.Body)

		error := &ErrorSchema{}

		err := json.Unmarshal(responseError, error)
		if err != nil {
			return errors.New(errorMessage + "\n\n" + string(errorMessage))
		}
		return errors.New(errorMessage + "\n\n" + error.ErrorMessage)
	}

	return nil
}
