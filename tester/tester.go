package tester

import (
	"fmt"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/proxy/lachesis"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

func PingNodesN(participants []*peers.Peer, p peers.PubKeyPeers, n uint64, serviceAddress string) {
	wg := new(sync.WaitGroup)
	fmt.Println("PingNodesN::participants: ", participants)
	fmt.Println("PingNodesN::p: ", p)
	for i := uint64(0); i < n; i++ {
		wg.Add(1)
		participant := participants[rand.Intn(len(participants))]
		node := p[participant.PubKeyHex]

		ipAddr, err := transact(*participant, node.ID, serviceAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
			fmt.Printf("Fatal error:\t\t\t%s\n", err.Error())
			if ipAddr != "" {
				fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", ipAddr, node)
			} else {
				fmt.Printf("Failed to ping:\t\t\tid=%d\n", node)
			}
			fmt.Printf("Failed to send transaction:\t%d\n\n", i)
		} else {
			fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", ipAddr, node)
			fmt.Printf("Last transaction sent:\t%d\n\n", i)
		}

		time.Sleep(1000 * time.Millisecond)
	}

	fmt.Println("Pinging stopped")

	wg.Wait()
}

func transact(target peers.Peer, nodeId int, proxyAddress string) (string, error) {
	addr := fmt.Sprintf("%s:%d", strings.Split(target.NetAddr, ":")[0], 9000)
	proxy := lachesis.NewSocketLachesisProxyClient(addr, 10 * time.Second)

	_, err := proxy.SubmitTx([]byte("oh hai"))
	// fmt.Println("Submitted tx, ack=", ack)  # `ack` is now `_`

	return "hi", err
}
