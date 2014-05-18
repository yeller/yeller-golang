package yeller

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type Client struct {
	ApiKey          string
	Version         string
	lastHostnameIdx int
	hostnames       []string
	httpClient      *http.Client
}

const CLIENT_VERSION = "yeller-golang: 0.0.1"

func NewClient(apiKey string) (client *Client) {
	rand.Seed(time.Now().UTC().UnixNano())

	hostnames := []string{
		"collector1.yellerapp.com",
		"collector2.yellerapp.com",
		"collector3.yellerapp.com",
		"collector4.yellerapp.com",
		"collector5.yellerapp.com",
	}

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
		Version:         CLIENT_VERSION,
		lastHostnameIdx: rand.Intn(len(hostnames)),
		hostnames:       hostnames,
		httpClient:      &httpClient,
	}
}

func (c *Client) Notify(note *ErrorNotification) error {
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
		return err
	}
	return nil
}

func (c *Client) tryNotifying(json []byte) error {
	url := "https://" + c.hostname() + "/" + c.ApiKey
	_, err := c.httpClient.Post(url, "application/json", bytes.NewReader(json))
	return err
}

func (c *Client) hostname() string {
	return c.hostnames[c.lastHostnameIdx]
}

func (c *Client) cycleHostname() {
	c.lastHostnameIdx = (c.lastHostnameIdx + 1) % len(c.hostnames)
}
