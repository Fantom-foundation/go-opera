package tester

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	_ "os"
	_ "sync"

	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/proxy/lachesis"
)

func PingNodesN(participants []*peers.Peer, p peers.PubKeyPeers, n uint64, serviceAddress string) {
	fmt.Println("PingNodesN::participants: ", participants)
	fmt.Println("PingNodesN::p: ", p)
	iteration := 0

	for {
		iteration++
		participant := participants[rand.Intn(len(participants))]
		node := p[participant.PubKeyHex]

		err := transact(*participant, node.ID, serviceAddress)

		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			fmt.Printf("Failed to send transaction:\t%d\n\n", iteration)
		} /*else {
			fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			fmt.Printf("Last transaction sent:\t%d\n\n", iteration)
		}*/
	}

	fmt.Println("Pinging stopped")
}

func transact(target peers.Peer, nodeId int, proxyAddress string) error {
	parts := strings.Split(target.NetAddr, ":")
	port, _ := strconv.Atoi(parts[1])
	addr := fmt.Sprintf("%s:%d", parts[0], port-3000)
	proxy := lachesis.NewSocketLachesisProxyClient(addr, 10*time.Second)

	// Ethereum txns are ~108 bytes. Bitcoin txns are ~250 bytes. We'll assume
	// our txns are ~120 bytes in size
	var msg [120]byte
	for i := 0; i < 100; i++ {
		// Send 10 txns to the server.
		_, err := proxy.SubmitTx(msg[:])
		if err != nil {
			return err
		}
	}
	// fmt.Println("Submitted tx, ack=", ack)  # `ack` is now `_`

	proxy.Close()
	return nil
}
