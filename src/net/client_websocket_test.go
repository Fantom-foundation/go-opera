package net

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

var c *ClientWebSocket

type ImpMessageListener struct{}

func (h *ImpMessageListener) OnMessage(v []byte, err error) {
	fmt.Println("OnMessage :", string(v), ", Err :", err)

	msg := Payload{}
	if err := json.Unmarshal(v, &msg); err != nil {
		fmt.Println("Failed to unmarshal msg :", err)
		return
	}

	switch msg.Counter {
	case 10:
		fmt.Println("Final Message received")
		c.Stop()
		return
	}

	pld := Payload{
		Counter: msg.Counter + 1,
		Time:    time.Now().Unix(),
	}

	b, _ := json.Marshal(pld)
	if err := c.SendMessage(b); err != nil {
		fmt.Println("Send message failed")
		return
	}
	fmt.Println("Send message succeed")
}

func TestClientWebSocket(t *testing.T) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGSTOP)
	signal.Notify(gracefulStop, syscall.SIGABRT)

	addr := "localhost:7777"
	path := "test"

	c = NewClientWebSocket()
	c.AddMessageListener(&ImpMessageListener{})

	go func() {
		if err := c.Connect(addr, path); err != nil {
			t.Error(err)
			return
		}
		fmt.Println("Websocket client has been connected")

		pld := Payload{
			Counter: 1,
			Time:    time.Now().Unix(),
		}

		b, _ := json.Marshal(pld)
		if err := c.SendMessage(b); err != nil {
			fmt.Println("Send message failed")
			return
		}
		fmt.Println("Send message succeed")
	}()

	<-gracefulStop

	c.Stop()

	fmt.Println("Websocket client has been disconnected")
}
