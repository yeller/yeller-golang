package yeller

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	ApiKey          string
	Environment     string
	Version         string
	roundtripMutex  *sync.Mutex
	lastHostnameIdx int
	hostnames       []string
	httpClient      *http.Client
	errorHandler    YellerErrorHandler
}

type YellerErrorHandler interface {
	HandleIOError(error) error
	HandleAuthError(error) error
}

const CLIENT_VERSION = "yeller-golang: 0.0.1"

func NewClient(apiKey string, env string, errorHandler YellerErrorHandler) (client *Client) {
	yellerHostnames := []string{
		"https://collector1.yellerapp.com",
		"https://collector2.yellerapp.com",
		"https://collector3.yellerapp.com",
		"https://collector4.yellerapp.com",
		"https://collector5.yellerapp.com",
	}
	return NewClientHostnames(apiKey, env, errorHandler, yellerHostnames)
}

func NewClientHostnames(apiKey string, env string, errorHandler YellerErrorHandler, hostnames []string) (client *Client) {
	// Set a timeout of 1 second before moving on to a different host
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			timeout := time.Duration(1 * time.Second)
			return net.DialTimeout(network, addr, timeout)
		},
	}
	httpClient := http.Client{Transport: &transport}

	return &Client{
		ApiKey:          apiKey,
		Environment:     env,
		Version:         CLIENT_VERSION,
		roundtripMutex:  &sync.Mutex{},
		lastHostnameIdx: randomHostnameIdx(hostnames),
		hostnames:       hostnames,
		httpClient:      &httpClient,
		errorHandler:    errorHandler,
	}
}

func (c *Client) Notify(note *ErrorNotification) error {
	c.cycleHostname()
	json, err := json.Marshal(note)
	if err != nil {
		return err
	}

	for _ = range c.hostnames {
		err = c.tryNotifying(json)
		if err == nil {
			break
		} else {
			c.cycleHostname()
		}
	}

	if err != nil {
		c.errorHandler.HandleIOError(err)
		return err
	}
	return nil
}

func (c *Client) tryNotifying(json []byte) error {
	url := c.hostname() + "/" + c.ApiKey
	response, err := c.httpClient.Post(url, "application/json", bytes.NewReader(json))
	if err != nil {
		return err
	}
	if response.StatusCode == 401 {
		authError := errors.New("Could not authenticate yeller client. Check your API key and that your subscription is active")
		c.errorHandler.HandleAuthError(authError)
		// Explictly return nil so that the client doesn't try to
		// round robin reporting this exception.
		return nil
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		// Don't report the error here, it gets
		// reported after we've round robined through clients
		return errors.New("Received a non 200 HTTP Code: " + response.Status)
	}
	return err
}

func (c *Client) hostname() string {
	return c.hostnames[c.lastHostnameIdx]
}

func (c *Client) cycleHostname() {
	c.roundtripMutex.Lock()
	defer func() {
		c.roundtripMutex.Unlock()
	}()
	c.lastHostnameIdx = (c.lastHostnameIdx + 1) % len(c.hostnames)
}

func randomHostnameIdx(hostnames []string) int {
	// Use a locally-scoped random source to avoid overwriting
	// global state.
	randSrc := rand.NewSource(time.Now().UTC().UnixNano())
	random := rand.New(randSrc)
	return random.Intn(len(hostnames))
}
