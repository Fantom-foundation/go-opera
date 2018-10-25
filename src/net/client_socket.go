package net

type MessageListener interface {
	OnMessage(msg []byte, err error)
}

type ClientSocket interface {
	Connect(addr string, path string) error
	AddMessageListener(listener MessageListener)
	SendMessage(msg []byte) error
	Stop() error
}
