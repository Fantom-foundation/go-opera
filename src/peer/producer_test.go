package peer_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/peer"
)

func TestProducer(t *testing.T) {
	ctx := context.Background()
	target := "1:2"
	limit := 2
	createFu := func(target string,
		timeout time.Duration) (peer.SyncClient, error) {
		return peer.NewClient(newRPCClient(t, nil, expSyncResponse))
	}
	producer := peer.NewProducer(limit, time.Second, createFu)

	// Test new connection.
	if producer.ConnLen(target) != 0 {
		t.Fatalf("expected %d, got %d", 0, producer.ConnLen(target))
	}

	cli, err := producer.Pop(target)
	if err != nil {
		t.Fatal(err)
	}

	resp := &peer.SyncResponse{}
	if err := cli.Sync(ctx, expSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}

	// Test reuse connection.
	producer.Push(target, cli)

	if producer.ConnLen(target) != 1 {
		t.Fatalf("expected %d, got %d", 1, producer.ConnLen(target))
	}

	cli, err = producer.Pop(target)
	if err != nil {
		t.Fatal(err)
	}

	resp = &peer.SyncResponse{}
	if err := cli.Sync(ctx, expSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}

	if producer.ConnLen(target) != 0 {
		t.Fatalf("expected %d, got %d", 0, producer.ConnLen(target))
	}

	// Test full pull.
	for i := 0; i < limit+1; i++ {
		producer.Push(target, cli)
	}

	if producer.ConnLen(target) != limit {
		t.Fatalf("expected %d, got %d", limit, producer.ConnLen(target))
	}

	// Test close producer.
	producer.Close()

	if _, err := producer.Pop(target); err != peer.ErrClientProducerStopped {
		t.Fatalf("expected %s, got %s", peer.ErrClientProducerStopped, err)
	}

	producer.Push(target, cli)

	if producer.ConnLen(target) != 0 {
		t.Fatalf("expected %d, got %d", 0, producer.ConnLen(target))
	}
}
