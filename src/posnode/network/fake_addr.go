package network

// Addr represents a fake network end point address.
type Addr string

/*
 * net.Addr implementation:
 */

// Network returns name of the network.
// It can be passed as the arguments to Dial().
func (a Addr) Network() string {
	return "fake"
}

// String returns string form of the network address.
func (a Addr) String() string {
	return string(a)
}
