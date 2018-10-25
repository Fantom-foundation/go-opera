package net

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

type Payload struct {
	Counter int   `json:"counter"`
	Time    int64 `json:"time"`
}

var ws *ServerWebSocket

type ImpOnConnectHook struct{}

func (h *ImpOnConnectHook) OnConnect(CID string, err error) {
	fmt.Println("New client with CID :", CID, "Err :", err)
}

type ImpOnMessageHook struct{}

func (h *ImpOnMessageHook) OnMessage(CID string, v []byte, err error) {
	fmt.Println("OnMessage CID :", CID, ", Msg :", string(v), ", Err :", err)

	msg := Payload{}
	if err := json.Unmarshal(v, &msg); err != nil {
		fmt.Println("Failed to unmarshal message :", err)
		return
	}

	pld := Payload{
		Counter: msg.Counter + 1,
		Time:    time.Now().Unix(),
	}

	b, _ := json.Marshal(pld)
	if err := ws.SendMessage(CID, b); err != nil {
		fmt.Println("Send reply failed")
		return
	}
	fmt.Println("Send reply succeed")
}

func TestNewWebSocket(t *testing.T) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGSTOP)
	signal.Notify(gracefulStop, syscall.SIGABRT)

	ws = NewServerWebSocket()
	ws.AddConnectHook(&ImpOnConnectHook{})
	ws.AddMessageHook(&ImpOnMessageHook{})

	go func() {
		fmt.Println("Websocket has been started")
		if err := ws.Listen("0.0.0.0:7777", "test"); err != nil {
			t.Error(err)
			return
		}
	}()

	<-gracefulStop

	ws.Stop(context.Background())

	fmt.Println("Websocket has been shutdown")
}
