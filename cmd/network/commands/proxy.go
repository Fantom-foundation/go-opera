package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tx string

// NewProxyCmd displays the version of babble being used
func NewProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Connect to a proxy",
		RunE:  connectProxy,
	}

	AddProxyFlags(cmd)

	return cmd
}

// Handler struct
type Handler struct {
	stateHash []byte
}

// CommitHandler Called when a new block is coming
// You must provide a method to compute the stateHash incrementally with incoming blocks
func (h *Handler) CommitHandler(block poset.Block) (stateHash []byte, err error) {
	hash := h.stateHash

	for _, tx := range block.Transactions() {
		fmt.Println(string(tx))
	}

	h.stateHash = crypto.Keccak256(append([][]byte{hash}, block.Transactions()...)...)

	return h.stateHash, nil
}

// SnapshotHandler Called when syncing with the network
func (h *Handler) SnapshotHandler(blockIndex int) (snapshot []byte, err error) {
	return []byte{}, nil
}

// RestoreHandler Called when syncing with the network
func (h *Handler) RestoreHandler(snapshot []byte) (stateHash []byte, err error) {
	return []byte{}, nil
}

// NewHandler constructor
func NewHandler() *Handler {
	return &Handler{}
}

func connectProxy(cmd *cobra.Command, args []string) error {
	i := config.Node

	babblePort := 1337
	proxyServPortStr := strconv.Itoa(babblePort + (i * 10) + 1)

	logger := logrus.New()

	logger.Level = logrus.InfoLevel

	appProxy, err := proxy.NewGrpcLachesisProxy("127.0.0.1:"+proxyServPortStr, logger)
	if err != nil {
		panic(err)
	}

	if len(tx) > 0 {
		if err := appProxy.SubmitTx([]byte(tx)); err != nil {
			panic(err)
		}

		return nil
	}

	if config.Stdin {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			if err := appProxy.SubmitTx(scanner.Bytes()); err != nil {
				panic(err)
			}
		}

		return nil
	}

	for {
		time.Sleep(time.Second)
	}
	return nil
}

// AddProxyFlags adds flags to the Run command
func AddProxyFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&config.Node, "node", config.Node, "Node index to connect to (starts from 0)")
	cmd.Flags().BoolVar(&config.Stdin, "stdin", config.Stdin, "Send some transactions from stdin")
	cmd.Flags().StringVar(&tx, "submit", tx, "Tx to submit and quit")
}
