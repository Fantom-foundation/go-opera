package reporter

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/gorilla/websocket"
	"github.com/recws-org/recws"
	"time"
)

type Config struct {
	Endpoint string
	Enabled  bool
	Verifier bool
}

type Reporter struct {
	Config
	ws        recws.RecConn
	p2pServer *p2p.Server
	stack     SingleStack
	queuedJob *WebSocketJob
}

func DefaultConfig() Config {
	return Config{
		Endpoint: "",
		Enabled:  false,
		Verifier: false,
	}
}

type WebSocketJob struct {
	PeerID   string
	BlockID  string
	Hash     string
	TimeDiff string
}

func NewReporter(cfg Config) *Reporter {
	return &Reporter{
		Config: cfg,
	}
}

func (x *Reporter) Send(blockID, hash, timeDiff string) {
	if x.p2pServer != nil {
		peerId := x.p2pServer.LocalNode().ID().String()[:6]
		x.stack.Push(&WebSocketJob{
			PeerID:   peerId,
			BlockID:  blockID,
			Hash:     hash,
			TimeDiff: timeDiff,
		})
	}
}

func (x *Reporter) sendDataOverWebSocket(peerID string, blockID string, hash string, timeDiff string) {
	// Prepare the data to be sent
	responseData := map[string]interface{}{
		"peer_id":   peerID,
		"block_id":  blockID,
		"hash":      hash,
		"time_diff": timeDiff,
	}

	// Marshal the response data to JSON
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		log.Error("JSON marshal error:", err)
		return
	}

	// Send the JSON response through the WebSocket
	log.Debug("Sending data to xenblocks Reporter", "data", responseData)
	if err := x.ws.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Error("Failed to send peer info to Reporter", "err", err)
	}
}

func (x *Reporter) worker() {
	for x.Enabled {
		job := x.stack.Pop()
		if job != nil {
			x.sendDataOverWebSocket(job.PeerID, job.BlockID, job.Hash, job.TimeDiff)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (x *Reporter) establishConnection() {
	log.Info("Establishing connection to Reporter")
	x.ws = recws.RecConn{
		NonVerbose: true,
	}
	x.ws.Dial(x.Endpoint, nil)
}

func (x *Reporter) Start(p2pServer *p2p.Server) *Reporter {
	if x.Endpoint == "" {
		return x
	}
	x.Enabled = true
	x.p2pServer = p2pServer
	x.establishConnection()
	x.stack = *NewSingleStack()
	go x.worker()
	return x
}

func (x *Reporter) Close() {
	x.Enabled = false
}
