package yeller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const ENV = "test"

func TestHostnameRotation(t *testing.T) {
	fakeYeller := NewFakeYeller(t, 5000, 5001, 5002)

	hostnames := []string{"localhost:5000", "localhost:5001", "localhost:5002"}
	client := NewClientHostnames("AN_API_KEY", ENV, NewStdErrErrorHandler(), hostnames)
	for _ = range hostnames {
		note := newErrorNotification(client, errors.New("an error"), nil)
		client.Notify(note)
	}

	fakeYeller.ShouldHaveReceivedRequestsOnPorts(map[int]int{
		5000: 1, 5001: 1, 5002: 1,
	})
}

type FakeYeller struct {
	Ports    []int
	requests []*http.Request
	servers  []*http.Server
	test     *testing.T
}

func NewFakeYeller(t *testing.T, ports ...int) *FakeYeller {
	fakeYeller := &FakeYeller{
		Ports: ports,
		test:  t,
	}

	for _, port := range ports {
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: fakeYeller,
		}
		go func() {
			if err := server.ListenAndServe(); err != nil {
				t.Error(err)
			}
		}()
		fakeYeller.servers = append(fakeYeller.servers, server)
	}

	return fakeYeller
}

func (f *FakeYeller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// XXX: Synchronize me bro
	f.requests = append(f.requests, r)
	w.WriteHeader(http.StatusOK)
}

func (f *FakeYeller) ShouldHaveReceivedRequestsOnPorts(exps map[int]int) {
	for _, r := range f.requests {
		hostPort := strings.Split(r.Host, ":")
		port, _ := strconv.Atoi(hostPort[len(hostPort)-1])
		times, ok := exps[port]
		if !ok {
			f.test.Errorf("received unexpected request on port %v", port)
		}
		if times <= 0 {
			f.test.Errorf("received too many requests on port %v", port)
		}
		exps[port] -= times - 1
	}
}
