package tester

import (
	"encoding/base64"
	"fmt"
	n "github.com/andrecronje/lachesis/net"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func PingNodesN(participants []n.Peer, n uint64) {
	txId := UniqueID{counter: 1}

	wg := new(sync.WaitGroup)
	for i := uint64(0); i < n; i++ {
		wg.Add(1)
		ticker := time.NewTicker(500 * time.Millisecond)

		go func() {
			for t := range ticker.C {
				rand.Seed(time.Now().Unix())
				participant := participants[rand.Intn(len(participants))]
				fmt.Printf("Pinging %s at %s\n", participant.NetAddr, t)
				sendTransact(participant, txId)
				fmt.Printf("Last transaction sent: %d\n", txId.Get()-1)
				// resp, err := http.Get("http://example.com/")
			}
		}()

		time.Sleep(1600 * time.Millisecond)
		ticker.Stop()
	}

	fmt.Println("Pinging stopped")

	wg.Wait()
}

func sendTransaction(target n.Peer) {
	ip := &layers.IPv4{
		SrcIP: GetOutboundIP(),
		DstIP: net.IP(target.NetAddr),
		// etc...
	}

	// TODO: Make shared counter for Tx #
	// TODO: Make shared counter for Node #
	payload := fmt.Sprintf("%s{\"method\":\"Lachesis.SubmitTx\",\"params\":[\"whatever\"],\"id\":\"whatever\"}",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Node%d Tx%d"))))

	net.Dial("tcp", ip.DstIP.String())

	buf := gopacket.NewSerializeBufferExpectedSize(len(payload), 0)
	opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
	err := ip.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.Bytes()) // prints out a byte slice containing
}

// https://stackoverflow.com/a/37382208
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func sendTransact(target n.Peer, txId UniqueID) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", target.NetAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		return
	}

	payload := fmt.Sprintf("%s{\"method\":\"Lachesis.SubmitTx\",\"params\":[\"whatever\"],\"id\":\"whatever\"}",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Node%d Tx%d", 900000000000000, txId.Get()))))

	_, err = conn.Write([]byte(payload))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		return
	}
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		return
	}
	fmt.Println(string(result))
}

type UniqueID struct {
	counter uint64
}

func (c *UniqueID) Get() uint64 {
	for {
		val := atomic.LoadUint64(&c.counter)
		if atomic.CompareAndSwapUint64(&c.counter, val, val+1) {
			return val
		}
	}
}
