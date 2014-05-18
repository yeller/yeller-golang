package yeller

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

type Client struct {
	ApiKey          string
	Version         string
	lastHostnameIdx int
	hostnames       []string
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

	return &Client{
		ApiKey:          apiKey,
		Version:         CLIENT_VERSION,
		lastHostnameIdx: rand.Intn(len(hostnames)),
		hostnames:       hostnames,
	}
}

func (c *Client) TryNotifying(note *ErrorNotification) error {
	json, err := json.Marshal(note)
	if err != nil {
		return err
	}

	for _ = range(c.hostnames) {
		url := "https://" + c.Hostname() + "/" + c.ApiKey
		_, err = http.Post(url, "application/json", bytes.NewReader(json))
		if err == nil {
			break
		} else {
			c.CycleHostname()
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Hostname() string {
	return c.hostnames[c.lastHostnameIdx]
}

func (c *Client) CycleHostname() {
	c.lastHostnameIdx = (c.lastHostnameIdx + 1) % len(c.hostnames)
}
