package fakenet

// Addr is a fake network address.
type Addr struct {
	AddressString string
	NetworkString string
}

// Network returns name of the network (for example, "tcp", "udp").
func (a Addr) Network() string {
	return a.NetworkString
}

// String returns a string form of address.
func (a Addr) String() string {
	return a.AddressString
}
