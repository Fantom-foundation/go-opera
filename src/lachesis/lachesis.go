package lachesis

import (
	"fmt"
	"net"
	"path/filepath"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/log"
	"github.com/Fantom-foundation/go-lachesis/src/node"
	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/service"
)

// Lachesis struct
type Lachesis struct {
	Config    *Config
	Node      *node.Node
	Transport peer.SyncPeer
	Store     poset.Store
	Peers     *peers.Peers
	Service   *service.Service
}

// NewLachesis constructor
func NewLachesis(config *Config) *Lachesis {
	engine := &Lachesis{
		Config: config,
	}

	return engine
}

func (l *Lachesis) initTransport() error {
	createCliFu := func(target string,
		timeout time.Duration) (peer.SyncClient, error) {

		rpcCli, err := peer.NewRPCClient(
			peer.TCP, target, time.Second, l.Config.ConnFunc)
		if err != nil {
			return nil, err
		}

		return peer.NewClient(rpcCli)
	}

	producer := peer.NewProducer(
		l.Config.MaxPool, l.Config.NodeConfig.TCPTimeout, createCliFu)
	backend := peer.NewBackend(
		peer.NewBackendConfig(), l.Config.Logger, net.Listen)
	if err := backend.ListenAndServe(peer.TCP, l.Config.BindAddr); err != nil {
		return err
	}
	l.Transport = peer.NewTransport(l.Config.Logger, producer, backend)
	return nil
}

func (l *Lachesis) initPeers() error {
	if !l.Config.LoadPeers {
		if l.Peers == nil {
			return fmt.Errorf("did not load peers but none was present")
		}

		return nil
	}

	peerStore := peers.NewJSONPeers(l.Config.DataDir)

	participants, err := peerStore.Peers()

	if err != nil {
		return err
	}

	if participants.Len() < 2 {
		return fmt.Errorf("peers.json should define at least two peers")
	}

	l.Peers = participants

	return nil
}

func (l *Lachesis) initStore() (err error) {
	if !l.Config.Store {
		l.Store = poset.NewInmemStore(l.Peers, l.Config.NodeConfig.CacheSize, &l.Config.PoSConfig)
		l.Config.Logger.Debug("created new in-mem store")
	} else {
		dbDir := l.Config.BadgerDir()
		l.Config.Logger.WithField("path", dbDir).Debug("Attempting to load or create database")
		l.Store, err = poset.LoadOrCreateBadgerStore(l.Peers, l.Config.NodeConfig.CacheSize, dbDir, &l.Config.PoSConfig)
		if err != nil {
			return
		}
	}

	if l.Store.NeedBootstrap() {
		l.Config.Logger.Debug("loaded store from existing database")
	} else {
		l.Config.Logger.Debug("created new store from blank database")
	}

	return
}

func (l *Lachesis) initKey() error {
	if l.Config.Key == nil {
		privKey, err := readKey(l.Config.DataDir)

		if err != nil {
			l.Config.Logger.Warn("Cannot read private key from file", err)

			privKey, err = Keygen(l.Config.DataDir)

			if err != nil {
				l.Config.Logger.Error("Cannot generate a new private key", err)

				return err
			}

			pem := fmt.Sprintf("0x%X", privKey.Public().Bytes())
			l.Config.Logger.Info("Created a new key:", pem)
		}

		l.Config.Key = privKey
	}

	return nil
}

func (l *Lachesis) initNode() error {
	key := l.Config.Key

	nodePub := fmt.Sprintf("0x%X", key.Public().Bytes())
	n, ok := l.Peers.ByPubKey[nodePub]

	if !ok {
		return fmt.Errorf("cannot find self pubkey in peers.json")
	}

	nodeID := n.ID

	l.Config.Logger.WithFields(logrus.Fields{
		"participants": l.Peers,
		"id":           nodeID,
	}).Debug("PARTICIPANTS")

	selectorArgs := node.SmartPeerSelectorCreationFnArgs{
		LocalAddr:    l.Config.BindAddr,
		GetFlagTable: nil,
	}
	l.Node = node.NewNode(
		&l.Config.NodeConfig,
		nodeID,
		key,
		l.Peers,
		l.Store,
		l.Transport,
		l.Config.Proxy,
		node.NewSmartPeerSelectorWrapper,
		selectorArgs,
		l.Config.BindAddr,
	)

	if err := l.Node.Init(); err != nil {
		return fmt.Errorf("failed to initialize node: %s", err)
	}

	return nil
}

func (l *Lachesis) initService() error {
	if l.Config.ServiceAddr != "" {
		l.Service = service.NewService(l.Config.ServiceAddr, l.Node, l.Config.Logger)
	}
	return nil
}

// Init initializes the lachesis node
func (l *Lachesis) Init() error {
	if l.Config.Logger == nil {
		l.Config.Logger = logrus.New()
		lachesis_log.NewLocal(l.Config.Logger, l.Config.LogLevel)
	}

	if err := l.initPeers(); err != nil {
		return err
	}

	if err := l.initStore(); err != nil {
		return err
	}

	if err := l.initTransport(); err != nil {
		return err
	}

	if err := l.initKey(); err != nil {
		return err
	}

	if err := l.initNode(); err != nil {
		return err
	}

	if err := l.initService(); err != nil {
		return err
	}

	return nil
}

// Run hosts the services for the lachesis node
func (l *Lachesis) Run() {
	if l.Service != nil {
		go l.Service.Serve()
	}
	l.Node.Run(true)
}

// Keygen generates a new key pair
func Keygen(datadir string) (*crypto.PrivateKey, error) {
	_, err := readKey(datadir)

	if err == nil {
		return nil, fmt.Errorf("another key already lives under %s", datadir)
	}

	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	if err := writeKey(datadir, privKey); err != nil {
		return nil, err
	}

	return privKey, nil
}

const (
	pemFile = "priv_key.pem"
)

func readKey(dataDir string) (*crypto.PrivateKey, error) {
	keyFile, err := os.Open(filepath.Join(dataDir, pemFile))
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.ReadKey(keyFile)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func writeKey(dataDir string, key *crypto.PrivateKey) error {
	keyFile, err := os.Create(filepath.Join(dataDir, pemFile))
	if err != nil {
		return err
	}

	if err := key.WriteTo(keyFile); err != nil {
		return err
	}

	return nil
}
