package yeller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type Client struct {
	Environment     string
	Version         string
	apiKey          string
	httpClient      *http.Client
	hostnames       []string
	lastHostnameIdx int
	queue           chan *ErrorNotification
}

const CLIENT_VERSION = "yeller-golang: 0.0.1"

func NewClient(apiKey string, env string) (client *Client) {
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

	client = &Client{
		Environment:     env,
		Version:         CLIENT_VERSION,
		apiKey:          apiKey,
		httpClient:      &httpClient,
		hostnames:       hostnames,
		lastHostnameIdx: rand.Intn(len(hostnames)),
		queue:           make(chan *ErrorNotification, 32),
	}

	go func() {
		for note := range client.queue {
			client.SyncNotify(note)
		}
	}()

	return client
}

func (c *Client) Notify(note *ErrorNotification) {
	c.queue <- note
}

func (c *Client) SyncNotify(note *ErrorNotification) {
	json, err := json.Marshal(note)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}
}

func (c *Client) tryNotifying(json []byte) error {
	url := "https://" + c.hostname() + "/" + c.apiKey
	_, err := c.httpClient.Post(url, "application/json", bytes.NewReader(json))
	return err
}

func (c *Client) hostname() string {
	return c.hostnames[c.lastHostnameIdx]
}

func (c *Client) cycleHostname() {
	c.lastHostnameIdx = (c.lastHostnameIdx + 1) % len(c.hostnames)
}
