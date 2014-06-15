package yeller

import (
	"errors"
	"net"
	"net/http"
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

type FakeYeller struct {
	Ports     []int
	requests  []*http.Request
	servers   []*http.Server
	test      *testing.T
	handler   func(f *FakeYeller, w http.ResponseWriter, r *http.Request)
	listeners []*net.Listener
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
	diagnosis := make(map[int]int)
	for _, r := range f.requests {
		hostPort := strings.Split(r.Host, ":")
		port, _ := strconv.Atoi(hostPort[len(hostPort)-1])
		times, ok := diagnosis[port]
		if !ok {
			diagnosis[port] = 0
			times = 0
		}
		diagnosis[port] += times + 1
	}
	for port, actualCount := range diagnosis {
		expectedCount, ok := exps[port]
		if !ok {
			f.test.Errorf("received unexpected request on port %v", port)
		}
		if actualCount != expectedCount {
			f.test.Errorf("received unexpected request count on port %v\ngot: %v\nexpected: %v", port, actualCount, expectedCount)
		}

	}
}

func (f *FakeYeller) shutdown() {
	for _, listener := range f.listeners {
		(*listener).Close()
	}
}
