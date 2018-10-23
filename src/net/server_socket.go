package net

import (
	"context"
)

type ClientConnectHook interface {
	OnConnect(CID string, err error)
}

type ClientMessageHook interface {
	OnMessage(CID string, msg []byte, err error)
}

type ServerSocket interface {
	Listen(addr string, path string) error
	AddConnectHook(hook ClientConnectHook)
	AddMessageHook(hook ClientMessageHook)
	SendMessage(CID string, msg []byte) error
	Stop(ctx context.Context) error
}
