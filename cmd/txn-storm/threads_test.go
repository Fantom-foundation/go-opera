package main

import (
	"testing"
	"time"
)

func TestThreads(t *testing.T) {
	tt := newThreads("http://18.222.120.223:4003", 1, 1, 3, 10, time.Minute)
	tt.Start()

	time.Sleep(5 * time.Second)
	tt.Stop()
}
