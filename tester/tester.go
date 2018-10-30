package tester

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	_ "os"
	_ "sync"

	"github.com/andrecronje/lachesis/src/peers"
	"github.com/sirupsen/logrus"
	"github.com/andrecronje/lachesis/src/proxy"
)

func PingNodesN(participants []*peers.Peer, p peers.PubKeyPeers, n uint64, logger *logrus.Logger) {
	proxies := make(map[int]*proxy.GrpcLachesisProxy)
	for _, participant := range participants {
		node := p[participant.PubKeyHex]
		host_port := strings.Split(node.NetAddr, ":")
		port, err := strconv.Atoi(host_port[1])
		addr := fmt.Sprintf("%s:%d", host_port[0], port-3000 /*9000*/)
		proxy, err := proxy.NewGrpcLachesisProxy(addr, logger)
		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to create WebsocketLachesisProxy:\t\t\t%s (id=%d)\n", participant.NetAddr, node.ID)
		}
		proxies[node.ID] = proxy
	}
	for iteration := uint64(0); iteration < n; iteration++ {
		participant := participants[rand.Intn(len(participants))]
		node := p[participant.PubKeyHex]

		_, err := transact(proxies[node.ID])

		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", participant.NetAddr, node.ID)
			fmt.Printf("Failed to send transaction:\t%d\n\n", iteration)
		} /*else {
			fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			fmt.Printf("Last transaction sent:\t%d\n\n", iteration)
		}*/
	}

	for _, proxy := range proxies {
		proxy.Close()
	}
	fmt.Println("Pinging stopped after ", n, " iterations")
}

func transact(proxy *proxy.GrpcLachesisProxy) (string, error) {

	// Ethereum txns are ~108 bytes. Bitcoin txns are ~250 bytes. We'll assume
	// our txns are ~120 bytes in size
	var msg [120]byte
	for i := 0; i < 10; i++ {
		// Send 10 txns to the server.
		err := proxy.SubmitTx(msg[:])
		if err != nil {
			return "", err
		}
	}
	// fmt.Println("Submitted tx, ack=", ack)  # `ack` is now `_`

	return "", nil
}
