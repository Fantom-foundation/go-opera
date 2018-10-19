package app

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type SocketAppProxyWebsocketServer struct {
	submitCh chan []byte
	logger   *logrus.Logger
}

func NewSocketAppProxyWebsocketServer(bindAddress string, logger *logrus.Logger) (*SocketAppProxyWebsocketServer, error) {
	server := &SocketAppProxyWebsocketServer{
		submitCh: make(chan []byte),
		logger:   logger,
	}

	go http.ListenAndServe(bindAddress, http.HandlerFunc(server.listen))

	return server, nil
}

func (p *SocketAppProxyWebsocketServer) listen(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to Upgrade", http.StatusInternalServerError)
		p.logger.WithField("error", err).Error("Failed to Upgrade")
		return
	}
	defer c.Close()

	for {
		_, tx, err := c.ReadMessage()
		if err != nil {
			p.logger.WithField("error", err).Error("websocket rx fail")
		}

		p.logger.Debug("SubmitTx")
		p.submitCh <- tx
	}
}
