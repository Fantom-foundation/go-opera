package net

// Message handler for server messages. Concrete implementations could
// be created using websockets, in-memory transport, UDP etc.
type MessageListener interface {
	OnMessage(msg []byte, err error)
}

// ClientSocket represents a client socket connection
type ClientSocket interface {
	Connect(addr string, path string) error
	AddMessageListener(listener MessageListener)
	SendMessage(msg []byte) error
	Stop() error
}
