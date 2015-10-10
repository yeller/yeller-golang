package yeller

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Options      map[string]interface{}
}

const (
	MAX_STACK_DEPTH = 256
)

var client *Client

func Start(apiKey string) {
	client = NewClient(apiKey, "production", "", NewStdErrErrorHandler())
}

func StartApplicationRoot(apiKey string, applicationRoot string) {
	client = NewClient(apiKey, "production", applicationRoot, NewStdErrErrorHandler())
}

func StartEnv(apiKey string, env string) {
	client = NewClient(apiKey, env, "", NewStdErrErrorHandler())
}

func StartEnvApplicationRoot(apiKey string, env string, applicationRoot string) {
	client = NewClient(apiKey, env, applicationRoot, NewStdErrErrorHandler())
}

func StartWithErrorHandler(apiKey string, env string, errorHandler YellerErrorHandler) {
	client = NewClient(apiKey, env, "", errorHandler)
}

func StartWithErrorHandlerApplicationRoot(apiKey string, applicationRoot string, errorHandler YellerErrorHandler) {
	client = NewClient(apiKey, "production", applicationRoot, errorHandler)
}

func StartWithErrorHandlerEnv(apiKey string, env string, errorHandler YellerErrorHandler) {
	client = NewClient(apiKey, env, "", errorHandler)
}

func StartWithErrorHandlerEnvApplicationRoot(apiKey string, env string, applicationRoot string, errorHandler YellerErrorHandler) {
	client = NewClient(apiKey, env, applicationRoot, errorHandler)
}

func StartWithClient(newClient *Client) {
	client = newClient
}

func Notify(appErr error) {
	NotifyInfo(appErr, nil)
}

func NotifyInfo(appErr error, info map[string]interface{}) {
	notification := newErrorNotification(client, appErr, info)
	client.Notify(notification)
}

func NotifyPanic(panicErr interface{}) {
	switch v := panicErr.(type) {
	case error:
		Notify(v)
	case string:
		Notify(errors.New(v))
	default:
		Notify(errors.New(fmt.Sprint(panicErr)))
	}
}

func NotifyPanicInfo(panicErr interface{}, info map[string]interface{}) {
	switch v := panicErr.(type) {
	case error:
		NotifyInfo(v, info)
	case string:
		NotifyInfo(errors.New(v), info)
	default:
		NotifyInfo(errors.New(fmt.Sprint(panicErr)), info)
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

	requestInfo := make(map[string]interface{})
	requestInfo["request-method"] = request.Method

	if len(request.Header["User-Agent"]) != 0 {
		requestInfo["user-agent"] = request.Header["User-Agent"][0]
	}

	if len(request.Header["Referer"]) != 0 {
		requestInfo["referrer"] = request.Header["Referer"][0]
	}

	newInfo["http-request"] = requestInfo

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
	fields := []interface{}{f.Filename, f.LineNumber, f.FunctionName, f.Options}
	return json.Marshal(fields)
}

func newErrorNotification(client *Client, appErr error, info map[string]interface{}) *ErrorNotification {
	if info == nil {
		info = make(map[string]interface{})
	}
	newErr := &ErrorNotification{
		Type:          "error",
		Message:       appErr.Error(),
		StackTrace:    applicationStackTrace(client),
		Host:          applicationHostname(),
		Environment:   client.Environment,
		CustomData:    info,
		ClientVersion: client.Version,
	}
	url, ok := info["url"].(string)
	if ok {
		newErr.Url = url
	}
	return newErr
}

func applicationHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func applicationStackTrace(client *Client) (stackTrace []StackFrame) {
	goroot := runtime.GOROOT()
	for i := 1; i <= MAX_STACK_DEPTH+1; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Ignore all stack frames coming from this package
		if strings.Contains(file, "github.com/yeller/yeller-golang") {
			continue
		}

		opts := make(map[string]interface{})
		if client.ProjectRoot != "" && strings.HasPrefix(file, client.ProjectRoot) {
			opts["in-app"] = true
		}
		file = strings.Replace(file, goroot, "", -1)
		frame := StackFrame{
			Filename:     file,
			LineNumber:   strconv.Itoa(line),
			FunctionName: functionName(pc),
			Options:      opts,
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
