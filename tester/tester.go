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
	proxies := make(map[int]*lachesis.WebsocketLachesisProxy)
	for _, participant := range participants {
		node := p[participant.PubKeyHex]
		addr := fmt.Sprintf("%s:%d", strings.Split(node.NetAddr, ":")[0], 9000)
		proxy, err := lachesis.NewWebsocketLachesisProxy(addr, nil, 10*time.Second, nil)
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

func transact(proxy *lachesis.WebsocketLachesisProxy) (string, error) {

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
