package tester

import (
	"fmt"
	"github.com/andrecronje/lachesis/net"
	"math/rand"
	"time"
)

func PingNodesContinuously(participants []net.Peer) {
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for t := range ticker.C {
			fmt.Printf("Pinging %s at %s\n", participants[rand.Intn(len(participants))].NetAddr, t)
			rand.Seed(time.Now().Unix())
			// resp, err := http.Get("http://example.com/")
		}
	}()
	time.Sleep(1600 * time.Millisecond)
	ticker.Stop()
	fmt.Println("Pinging stopped")
}
