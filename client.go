package yeller

import (
	"math/rand"
	"time"
)

type Client struct {
	ApiKey          string
	LastHostnameIdx int
	Hostnames       []string
	Version         string
}

const CLIENT_VERSION = "yeller-golang: 0.0.1"

func NewClient(apiKey string) *Client {
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
		LastHostnameIdx: rand.Intn(len(hostnames)),
		Hostnames:       hostnames,
		Version:         CLIENT_VERSION,
	}
}

func (c *Client) Hostname() string {
	return c.Hostnames[c.LastHostnameIdx]
}

func (c *Client) NextHostname() string {
	c.LastHostnameIdx++
	return c.Hostname()
}
