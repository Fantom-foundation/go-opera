package tester

import (
	"fmt"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/proxy/lachesis"
	"math/rand"
	_ "os"
	"strings"
	"strconv"
	_ "sync"
	"time"
)

func PingNodesN(participants []*peers.Peer, p peers.PubKeyPeers, n uint64, serviceAddress string) {
	// wg := new(sync.WaitGroup)
	fmt.Println("PingNodesN::participants: ", participants)
	fmt.Println("PingNodesN::p: ", p)
	// for i := uint64(0); i < n; i++ {
	
	// delay := 1 * time.Millisecond
	iteration := 0

	for {
		iteration++
		// wg.Add(1)
		participant := participants[rand.Intn(len(participants))]
		node := p[participant.PubKeyHex]

		// fmt.Println("transacting", iteration)
		err := transact(*participant, node.ID, serviceAddress)
		// fmt.Println("transact done", iteration)
		// Simple exponential backoff
		// if err == nil {
		// 	if delay > 1 * time.Millisecond {
		// 		delay = time.Duration(float64(delay) * 0.9)
		// 	}
		// } else {
		// 	delay = time.Duration(float64(delay) * 1.3)
		// }
		
		
		if err != nil {
			// fmt.Fprintf(os.Stderr, "Error injecting events: %s\n", err.Error())
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			fmt.Printf("Failed to send transaction:\t%d\n\n", iteration)
		} else {
			// fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			// fmt.Printf("Last transaction sent:\t%d\n\n", iteration)
		}

		// if iteration % 1000 == 0 {
		// 	fmt.Println("delay", delay)
		// }
		// time.Sleep(delay)
	}

	fmt.Println("Pinging stopped")

	// wg.Wait()
}

func transact(target peers.Peer, nodeId int, proxyAddress string) (error) {
	parts := strings.Split(target.NetAddr, ":")
	port, _ := strconv.Atoi(parts[1])
	addr := fmt.Sprintf("%s:%d", parts[0], port - 3000)
	// addr := "127.0.0.1:9000"
	proxy := lachesis.NewSocketLachesisProxyClient(addr, 10 * time.Second)

	// Ethereum txns are ~108 bytes. Bitcoin txns are ~250 bytes. We'll assume
	// our txns are ~120 bytes in size
	var msg [120]byte;

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
