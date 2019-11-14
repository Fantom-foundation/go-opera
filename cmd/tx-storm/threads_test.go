package main

import (
	"testing"
	"time"
)

func TestThreads(t *testing.T) {
	const url = "http://127.0.0.1:4002"

	tt := newThreads(url, 0, 1, 3, 0, 100)
	tt.Start()

	time.Sleep(5 * time.Second)
	tt.Stop()
}
