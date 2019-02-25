package peer

import (
	"sync"
	"time"
)

// ClientProducer is an interface representing methods for producer of sync
// clients.
type ClientProducer interface {
	Pop(target string) (SyncClient, error)
	Push(target string, client SyncClient)
	Close()
}

// Producer creates new sync clients. Stores a limited number of clients in
// a pool for reuse.
type Producer struct {
	createFunc CreateSyncClientFunc
	poolSize   int
	timeout    time.Duration

	mtx      sync.Mutex
	pool     map[string][]SyncClient
	shutdown bool
}

// NewProducer creates new producer of sync clients.
func NewProducer(poolSize int, connectTimeout time.Duration,
	createClientFunc CreateSyncClientFunc) *Producer {
	return &Producer{
		createFunc: createClientFunc,
		poolSize:   poolSize,
		timeout:    connectTimeout,
		pool:       make(map[string][]SyncClient),
	}
}

// Pop creates a new connection for a target or re-uses an existing connection.
func (p *Producer) Pop(target string) (SyncClient, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.shutdown {
		return nil, ErrClientProducerStopped
	}

	clients := p.pool[target]
	if len(clients) != 0 {
		var cli SyncClient
		num := len(clients)

		cli, clients[num-1] = clients[num-1], nil
		p.pool[target] = clients[:num-1]
		return cli, nil
	}

	return p.createFunc(target, p.timeout)
}

// Push saves a connection in a pool.
func (p *Producer) Push(target string, client SyncClient) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.shutdown || len(p.pool[target]) >= p.poolSize {
		if err := client.Close(); err != nil {
			panic(err)
		}
		return
	}

	p.pool[target] = append(p.pool[target], client)
}

// Close closes a producer.
func (p *Producer) Close() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.shutdown {
		return
	}

	p.shutdown = true

	for target := range p.pool {
		for k := range p.pool[target] {
			if err := p.pool[target][k].Close(); err != nil {
				panic(err)
			}
		}
	}
	p.pool = nil
}

// ConnLen returns the number of connections in a pool for a specific target.
func (p *Producer) ConnLen(target string) int {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	return p.connLen(target)
}

func (p *Producer) connLen(target string) int {
	if p.shutdown {
		return 0
	}

	return len(p.pool[target])
}
