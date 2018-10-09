package net

// LSocket represents a socket connection
type LSocket interface {
	Connect() error
	FireEvent(v interface{}, topic string) error
	Listen(topic string) error
	Disconnect()
}
