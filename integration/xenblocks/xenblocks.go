package xenblocks

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
}

type Xenblocks struct {
	Config
	ws        recws.RecConn
	p2pServer *p2p.Server
	stack     SingleStack
	queuedJob *WebSocketJob
}

func DefaultConfig() Config {
	return Config{
		Endpoint: "",
	}
}

type WebSocketJob struct {
	PeerID   string
	BlockID  string
	Hash     string
	TimeDiff string
}

func (x *Xenblocks) Send(blockID, hash, timeDiff string) {
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

func (x *Xenblocks) sendDataOverWebSocket(peerID string, blockID string, hash string, timeDiff string) {
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
	log.Info("Sending data to XenBlocks", "data", responseData)
	if err := x.ws.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Error("Failed to send peer info to xenblocks", "err", err)
	}
}

func (x *Xenblocks) worker() {
	for x.Enabled {
		job := x.stack.Pop()
		if job != nil {
			x.sendDataOverWebSocket(job.PeerID, job.BlockID, job.Hash, job.TimeDiff)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (x *Xenblocks) establishConnection() {
	log.Info("Establishing connection to XenBlocks")
	x.ws = recws.RecConn{
		NonVerbose: true,
	}
	x.ws.Dial(x.Endpoint, nil)
}

func (x *Xenblocks) Start(p2pServer *p2p.Server) *Xenblocks {
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

func (x *Xenblocks) Stop() {
	x.Enabled = false
}
