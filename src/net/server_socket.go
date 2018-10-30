package net

import (
	"context"
)

// ClientConnectHook provides an interface to handle when a client application connects to the server. 
type ClientConnectHook interface {
	OnConnect(CID string, err error)
}

// ClientMessageHook provides an interface to handle messages from the client application.
type ClientMessageHook interface {
	OnMessage(CID string, msg []byte, err error)
}

// ServerSocket represents a server socket connection. 
type ServerSocket interface {
	Listen(addr string, path string) error
	AddConnectHook(hook ClientConnectHook)
	AddMessageHook(hook ClientMessageHook)
	SendMessage(CID string, msg []byte) error
	Stop(ctx context.Context) error
}
