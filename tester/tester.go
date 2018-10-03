package tester

import (
	"encoding/base64"
	"fmt"
	lachesisNet "github.com/andrecronje/lachesis/net"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func PingNodesN(participants []lachesisNet.Peer, p map[string]int, n uint64, proxyAddress string) {
	txId := UniqueID{counter: 1}

	wg := new(sync.WaitGroup)
	fmt.Println("PingNodesN::participants: ", participants)
	fmt.Println("PingNodesN::p: ", p)
	for i := uint64(0); i < n; i++ {
		wg.Add(1)
		participant := participants[rand.Intn(len(participants))]
		nodeId := p[participant.NetAddr]

		ipAddr, err := transact(participant, nodeId, txId, proxyAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
			fmt.Printf("Fatal error:\t\t\t%s\n", err.Error())
			if ipAddr != "" {
				fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", ipAddr, nodeId)
			} else {
				fmt.Printf("Failed to ping:\t\t\tid=%d\n", nodeId)
			}
			fmt.Printf("Failed to send transaction:\t%d\n\n", txId.Get()-1)
		} else {
			fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", ipAddr, nodeId)
			fmt.Printf("Last transaction sent:\t%d\n\n", txId.Get()-1)
		}

		time.Sleep(1600 * time.Millisecond)
	}

	fmt.Println("Pinging stopped")

	wg.Wait()
}

func sendTransaction(target lachesisNet.Peer) {
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

func transact(target lachesisNet.Peer, nodeId int, txId UniqueID, proxyAddress string) (string, error) {
	// rpcClient := jsonrpc.NewClient(proxyAddress)
	// rpcClient.Call("createPerson", "Alex", 33, "Germany")
	// generates body: {"method":"createPerson","params":["Alex",33,"Germany"],"id":0,"jsonrpc":"2.0"}

	tcpAddr, err := net.ResolveTCPAddr("tcp4",
		fmt.Sprintf("%s:%d", strings.Split(target.NetAddr, ":")[0], 9000))
	if err != nil {
		return "", err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return "", err
	}

	payload := fmt.Sprintf("%s{\"method\":\"Lachesis.SubmitTx\",\"params\":[\"whatever\"],\"id\":\"whatever\"}",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Node%d Tx%d", nodeId, txId.Get()))))

	_, err = conn.Write([]byte(payload))

	if err != nil {
		return "", err
	}
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", err
	}
	fmt.Println(string(result))
	return tcpAddr.String(), err
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
