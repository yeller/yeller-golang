package yeller

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type ErrorNotification struct {
	Type          string                 `json:"type"`
	Message       string                 `json:"message"`
	StackTrace    []StackFrame           `json:"stacktrace"`
	Url           string                 `json:"url"`
	Host          string                 `json:"host"`
	Environment   string                 `json:"application-environment"`
	CustomData    map[string]interface{} `json:"custom-data"`
	Location      string                 `json:"location"`
	ClientVersion string                 `json:"client-version"`
}

type StackFrame struct {
	Filename     string
	LineNumber   string
	FunctionName string
}

const (
	MAX_STACK_DEPTH = 256
)

var client *Client

func Start(apiKey string) {
	client = NewClient(apiKey, "production")
}

func StartEnv(apiKey string, env string) {
	client = NewClient(apiKey, env)
}

func Notify(appErr error) {
	NotifyInfo(appErr, make(map[string]interface{}))
}

func NotifyInfo(appErr error, info map[string]interface{}) {
	notification := newErrorNotification(appErr, info)
	err := client.Notify(notification)
	if err != nil {
		log.Println(err)
	}
}

func NotifyHTTP(appErr error, request http.Request) {
	info := make(map[string]interface{})
	NotifyHTTPInfo(appErr, request, info)
}

func NotifyHTTPInfo(appErr error, request http.Request, info map[string]interface{}) {
	// we have to copy the values out of the
	// map because we're about to mutate the map
	// and we don't want to mutate user provided data
	newInfo := make(map[string]interface{})

	formErr := request.ParseForm()
	if formErr != nil {
		newInfo["Params"] = request.Form
	}
	newInfo["Cookies"] = getCookies(request)
	newInfo["url"] = request.URL

	for k, v := range info {
		newInfo[k] = v
	}
	NotifyInfo(appErr, newInfo)
}

func getCookies(request http.Request) map[string]interface{} {
	cookies := make(map[string]interface{})
	for _, cookie := range request.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return cookies
}

func (f StackFrame) MarshalJSON() ([]byte, error) {
	fields := []string{f.Filename, f.LineNumber, f.FunctionName}
	return json.Marshal(fields)
}

func newErrorNotification(appErr error, info map[string]interface{}) *ErrorNotification {
	newErr := ErrorNotification{
		Type:          "error",
		Message:       appErr.Error(),
		StackTrace:    applicationStackTrace(),
		Host:          applicationHostname(),
		Environment:   client.Environment,
		CustomData:    info,
		ClientVersion: client.Version,
	}
	url, ok := info["url"].(string)
	if ok {
		newErr.Url = url
	}
	return &newErr
}

func applicationHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func applicationStackTrace() (stackTrace []StackFrame) {
	for i := 1; i <= MAX_STACK_DEPTH+1; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Ignore all stack frames coming from this package
		if strings.Contains(file, "github.com/yeller/yeller-golang") {
			continue
		}

		frame := StackFrame{
			Filename:     file,
			LineNumber:   strconv.Itoa(line),
			FunctionName: functionName(pc),
		}
		stackTrace = append(stackTrace, frame)
	}
	return stackTrace
}

func functionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "???"
	}
	return fn.Name()
}
