package net

import (
	"context"
	"fmt"
	"github.com/andrecronje/lachesis/src/utils"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"net/http"
)

type ServerWebSocket struct {
	s            *http.Server
	connectHooks []ClientConnectHook
	messageHooks []ClientMessageHook
	clients      map[string]*websocket.Conn
}

func NewServerWebSocket() *ServerWebSocket {
	return &ServerWebSocket{
		clients: map[string]*websocket.Conn{},
	}
}

func (ws *ServerWebSocket) Listen(addr string, path string) error {
	return ws.run(addr, path)
}

func (ws *ServerWebSocket) AddConnectHook(hook ClientConnectHook) {
	ws.connectHooks = append(ws.connectHooks, hook)
}

func (ws *ServerWebSocket) AddMessageHook(hook ClientMessageHook) {
	ws.messageHooks = append(ws.messageHooks, hook)
}

func (ws *ServerWebSocket) SendMessage(CID string, v []byte) error {
	return ws.clients[CID].WriteMessage(websocket.TextMessage, v)
}

func (ws *ServerWebSocket) Stop(ctx context.Context) error {
	return ws.s.Shutdown(ctx)
}

func (ws *ServerWebSocket) run(addr string, path string) error {
	ws.s = &http.Server{
		Addr:    addr,
		Handler: ws.registerRoutes(path),
	}
	return ws.s.ListenAndServe()
}

func (ws *ServerWebSocket) registerRoutes(path string) http.Handler {
	h := chi.NewMux()
	h.Use(middleware.Logger)

	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	h.HandleFunc(fmt.Sprintf("/%s", path), func(w http.ResponseWriter, r *http.Request) {
		uuid := utils.UUID()
		up := websocket.Upgrader{}
		up.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			for _, hook := range ws.connectHooks {
				hook.OnConnect(uuid, err)
			}
			return
		}
		for _, hook := range ws.connectHooks {
			hook.OnConnect(uuid, nil)
		}
		ws.clients[uuid] = conn
		go ws.readMessage(uuid, conn)
	})
	return h
}

func (ws *ServerWebSocket) readMessage(CID string, conn *websocket.Conn) {
	for {
		t, m, err := conn.ReadMessage()
		if err != nil {
			delete(ws.clients, CID)
			return
		}

		switch t {
		case websocket.TextMessage:
			ws.notifyMessageReceive(CID, m, err)
		case websocket.BinaryMessage:
			//ws.notifyMessageReceive(CID, m, err)
		case websocket.CloseMessage:
			delete(ws.clients, CID)
		case websocket.PingMessage:
			ws.clients[CID].WriteMessage(websocket.PongMessage, nil)
		case websocket.PongMessage:
			ws.clients[CID].WriteMessage(websocket.PingMessage, nil)
		default:
			fmt.Println("Unknown message with type :", t)
		}
	}
}

func (ws *ServerWebSocket) notifyMessageReceive(CID string, v []byte, err error) {
	for _, hook := range ws.messageHooks {
		hook.OnMessage(CID, v, err)
	}
}
