package yeller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type ErrorNotification struct {
	Type          string                 `json:"type"`
	Message       string                 `json:"message"`
	StackTrace    []string               `json:"stacktrace"`
	Url           string                 `json:"url"`
	Host          string                 `json:"host"`
	CustomData    map[string]interface{} `json:"custom-data"`
	Location      string                 `json:"location"`
	ClientVersion string                 `json:"client-version"`
}

const CLIENT_VERSION = "yeller-golang: 0.0.1"

var apiKey string

func Start(newApiKey string) {
	apiKey = newApiKey
}

func Notify(err error) {
	notification := NewErrorNotification(err)

	json, err := json.Marshal(notification)
	if err != nil {
		log.Println(err)
		return
	}

	url := "https://collector1.yellerapp.com/" + apiKey
	resp, err := http.Post(url, "application/json", bytes.NewReader(json))
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(resp)
}

func NewErrorNotification(err error) *ErrorNotification {
	return &ErrorNotification{
		Type: "error",
		Message: err.Error(),
		ClientVersion: CLIENT_VERSION,
	}
}
