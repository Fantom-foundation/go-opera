package tester

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	_ "os"
	_ "sync"

	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/proxy/socket/lachesis"
)

func PingNodesN(participants []*peers.Peer, p peers.PubKeyPeers, n uint64, serviceAddress string) {
	for iteration := uint64(0); iteration < n; iteration++ {
		participant := participants[rand.Intn(len(participants))]
		node := p[participant.PubKeyHex]

		_, err := transact(*participant, node.ID, serviceAddress)

		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", participant.NetAddr, node.ID)
			fmt.Printf("Failed to send transaction:\t%d\n\n", iteration)
		} /*else {
			fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			fmt.Printf("Last transaction sent:\t%d\n\n", iteration)
		}*/
	}

	fmt.Println("Pinging stopped")
}

func transact(target peers.Peer, nodeId int, proxyAddress string) (string, error) {
	addr := fmt.Sprintf("%s:%d", strings.Split(target.NetAddr, ":")[0], 9000)
	proxy := lachesis.NewSocketLachesisProxyClientWebsocket(addr, 10*time.Second)

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

	proxy.Close()
	return "", nil
}
