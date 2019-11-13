package main

import (
	"testing"
	"time"
)

func ThreadsExample(t *testing.T) {
	const url = "http://127.0.0.1:4002"

	tt := newThreads(url, 1, 1, 3, 10, time.Minute)
	tt.Start()

	time.Sleep(5 * time.Second)
	tt.Stop()
}
