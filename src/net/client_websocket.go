package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
)

type ClientWebSocket struct {
	conn             *websocket.Conn
	messageListeners []MessageListener
}

func NewClientWebSocket() *ClientWebSocket {
	return &ClientWebSocket{}
}

func (ws *ClientWebSocket) Connect(addr string, path string) error {
	return ws.run(addr, path)
}

func (ws *ClientWebSocket) AddMessageListener(listener MessageListener) {
	ws.messageListeners = append(ws.messageListeners, listener)
}

func (ws *ClientWebSocket) SendMessage(v []byte) error {
	return ws.conn.WriteMessage(websocket.TextMessage, v)
}

func (ws *ClientWebSocket) Stop() error {
	return ws.conn.Close()
}

func (ws *ClientWebSocket) run(addr string, path string) error {
	u := url.URL{
		Scheme: "ws",
		Host:   addr,
		Path:   path,
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	ws.conn = c
	go ws.readMessage()
	return nil
}

func (ws *ClientWebSocket) readMessage() {
	for {
		t, m, err := ws.conn.ReadMessage()
		if err != nil {
			return
		}

		switch t {
		case websocket.TextMessage:
			ws.notifyMessageReceive(m, err)
		case websocket.BinaryMessage:
			//ws.notifyMessageReceive(CID, m, err)
		case websocket.CloseMessage:
			ws.conn.Close()
		case websocket.PingMessage:
			ws.conn.WriteMessage(websocket.PongMessage, nil)
		case websocket.PongMessage:
			ws.conn.WriteMessage(websocket.PingMessage, nil)
		default:
			fmt.Println("Unknown message with type :", t)
		}
	}
}

func (ws *ClientWebSocket) notifyMessageReceive(v []byte, err error) {
	for _, hook := range ws.messageListeners {
		hook.OnMessage(v, err)
	}
}
