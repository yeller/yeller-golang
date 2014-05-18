package yeller

import (
	"os"
	"strconv"
	"testing"
	"time"
)

const API_KEY = "YOUR_API_KEY"

func TestYeller(t *testing.T) {
	StartEnv(API_KEY, "staging")
	for i := 0; i < 15; i++ {
		_, err := os.Open("filename" + strconv.Itoa(i) + ".ext")
		if err != nil {
			Notify(err)
		}
	}
	time.Sleep(30 * time.Second)
}
