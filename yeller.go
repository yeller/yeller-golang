package yeller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
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
	CLIENT_VERSION  = "yeller-golang: 0.0.1"
	MAX_STACK_DEPTH = 256
)

var apiKey string

func Start(newApiKey string) {
	apiKey = newApiKey
}

func Notify(err error) {
	notification := newErrorNotification(err)

	json, err := json.Marshal(notification)
	if err != nil {
		log.Println(err)
		return
	}

	url := "https://collector1.yellerapp.com/" + apiKey
	_, err = http.Post(url, "application/json", bytes.NewReader(json))
	if err != nil {
		log.Println(err)
		return
	}
}

func (f StackFrame)MarshalJSON() ([]byte, error) {
	fields := []string{f.Filename, f.LineNumber, f.FunctionName}
	return json.Marshal(fields)
}

func newErrorNotification(err error) *ErrorNotification {
	return &ErrorNotification{
		Type:          "error",
		Message:       err.Error(),
		StackTrace:    applicationStackTrace(),
		ClientVersion: CLIENT_VERSION,
	}
}

func applicationStackTrace() (stackTrace []StackFrame) {
	for i := 1; i <= MAX_STACK_DEPTH + 1; i++ {
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
