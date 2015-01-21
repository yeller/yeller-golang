package yeller

import (
	"errors"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const ENV = "test"

func TestSendingExceptionsToMultipleServers(t *testing.T) {
	fakeYellerHandler := func(f *FakeYeller, w http.ResponseWriter, r *http.Request) {
		f.requests = append(f.requests, r)
		w.WriteHeader(http.StatusOK)
	}
	fakeYeller := NewFakeYeller(t, fakeYellerHandler, 5000, 5001, 5002)

	hostnames := []string{"http://localhost:5000", "http://localhost:5001", "http://localhost:5002"}
	client := NewClientHostnames("AN_API_KEY", ENV, NewPanicErrorHandler(), hostnames)
	for _ = range hostnames {
		note := newErrorNotification(client, errors.New("an error"), nil)
		client.Notify(note)
	}

	fakeYeller.ShouldHaveReceivedRequestsOnPorts(map[int]int{
		5000: 1, 5001: 1, 5002: 1,
	})
	fakeYeller.shutdown()
}

func TestRoundTrippingFailingServers(t *testing.T) {
	fakeYellerHandler := func(f *FakeYeller, w http.ResponseWriter, r *http.Request) {
		hostPort := strings.Split(r.Host, ":")
		port, _ := strconv.Atoi(hostPort[len(hostPort)-1])
		if port == 5000 {
			f.requests = append(f.requests, r)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

	}
	fakeYeller := NewFakeYeller(t, fakeYellerHandler, 5000, 5001, 5002)

	hostnames := []string{"http://localhost:5000", "http://localhost:5001", "http://localhost:5002"}
	client := NewClientHostnames("AN_API_KEY", ENV, NewPanicErrorHandler(), hostnames)
	for _ = range hostnames {
		note := newErrorNotification(client, errors.New("an error"), nil)
		client.Notify(note)
	}

	fakeYeller.ShouldHaveReceivedRequestsOnPorts(map[int]int{
		5000: 3, 5001: 0, 5002: 0,
	})
	fakeYeller.shutdown()
}

type FakeYeller struct {
	Ports              []int
	requests           []*http.Request
	errorNotifications []*ErrorNotification
	servers            []*http.Server
	test               *testing.T
	handler            func(f *FakeYeller, w http.ResponseWriter, r *http.Request)
	listeners          []*net.Listener
}

func NewFakeYeller(t *testing.T, handler func(f *FakeYeller, w http.ResponseWriter, r *http.Request), ports ...int) *FakeYeller {
	fakeYeller := &FakeYeller{
		Ports:   ports,
		test:    t,
		handler: handler,
	}

	for _, port := range ports {
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: fakeYeller,
		}
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			t.Error(err)
		}
		fakeYeller.listeners = append(fakeYeller.listeners, &listener)
		go server.Serve(listener)
		fakeYeller.servers = append(fakeYeller.servers, server)
	}

	return fakeYeller
}

func (f *FakeYeller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// XXX: Synchronize me bro
	f.handler(f, w, r)
}

func (f *FakeYeller) ShouldHaveReceivedRequestsOnPorts(exps map[int]int) {
	actual := make(map[int]int)
	for _, r := range f.requests {
		hostPort := strings.Split(r.Host, ":")
		port, _ := strconv.Atoi(hostPort[len(hostPort)-1])
		_, ok := actual[port]
		if !ok {
			actual[port] = 0
		}
		actual[port] += 1
	}
	for port, actualCount := range actual {
		expectedCount, ok := exps[port]
		if !ok {
			f.test.Errorf("received unexpected request on port %v", port)
		}
		if actualCount != expectedCount {
			f.test.Errorf("received unexpected request count on port %v\ngot: %v\nexpected: %v", port, actualCount, expectedCount)
		}

	}
}

func (f *FakeYeller) ShouldHaveReceivedRequestWithInfo(exps map[string]interface{}) {
	received := false
	for _, r := range f.errorNotifications {
		for k, v := range exps {
			if reflect.DeepEqual(r.CustomData[k], v) {
				received = true
			}
		}
	}
	if !received {
		f.test.Errorf("didn't receive request with matching info, %s", f.errorNotifications)
	}
}

func (f *FakeYeller) shutdown() {
	for _, listener := range f.listeners {
		(*listener).Close()
	}
}
