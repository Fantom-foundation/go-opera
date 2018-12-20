package net

import (
	"testing"
)

func TestInmemTransport(t *testing.T) {

	// Transport 1 is consumer
	_, trans1 := NewInmemTransport("")
	defer trans1.Close()

	// Transport 2 makes outbound request
	_, trans2 := NewInmemTransport("")
	defer trans2.Close()

	testTransportImplementation(t, trans1, trans2)
}
