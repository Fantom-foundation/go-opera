package net

import (
	"context"
)

type OnConnectHook interface {
	OnConnect(CID string, err error)
}

type OnMessageHook interface {
	OnMessage(CID string, msg []byte, err error)
}

type ServerSocket interface {
	Listen(addr string, path string) error
	AddConnectHook(hook OnConnectHook)
	AddMessageHook(hook OnMessageHook)
	SendMessage(CID string, msg []byte) error
	Stop(ctx context.Context) error
}
