package yeller

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestSendingErrorsWithHTTPInfo(t *testing.T) {
	fakeYellerHandler := func(f *FakeYeller, w http.ResponseWriter, r *http.Request) {
		f.requests = append(f.requests, r)
		decoder := json.NewDecoder(r.Body)
		var info ErrorNotification
		decoder.Decode(&info)
		f.errorNotifications = append(f.errorNotifications, &info)
		w.WriteHeader(http.StatusOK)
	}
	fakeYeller := NewFakeYeller(t, fakeYellerHandler, 5000)

	hostnames := []string{"http://localhost:5000", "http://localhost:5001", "http://localhost:5002"}
	client := NewClientHostnames("AN_API_KEY", ENV, NewPanicErrorHandler(), hostnames)
	StartWithClient(client)
	req, _ := http.NewRequest("GET", "", nil)
	NotifyHTTP(errors.New("an error"), *req)
	httpInfo := make(map[string]interface{})
	httpInfo["request-method"] = "GET"
	fakeYeller.ShouldHaveReceivedRequestWithInfo(map[string]interface{}{
		"http-request": httpInfo,
	})
}
