package lachesis

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/net"
	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/peers"
	h "github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/service"
	"github.com/sirupsen/logrus"
)

type Lachesis struct {
	Config    *LachesisConfig
	Node      *node.Node
	Transport net.Transport
	Store     h.Store
	Peers     *peers.Peers
	Service   *service.Service
}

func NewLachesis(config *LachesisConfig) *Lachesis {
	engine := &Lachesis{
		Config: config,
	}

	return engine
}

func (b *Lachesis) initTransport() error {
	transport, err := net.NewTCPTransport(
		b.Config.BindAddr,
		nil,
		b.Config.MaxPool,
		b.Config.NodeConfig.TCPTimeout,
		b.Config.Logger,
	)

	if err != nil {
		return err
	}

	b.Transport = transport

	return nil
}

func (b *Lachesis) initPeers() error {
	if !b.Config.LoadPeers {
		if b.Peers == nil {
			return fmt.Errorf("Did not load peers but none was present")
		}

		return nil
	}

	peerStore := peers.NewJSONPeers(b.Config.DataDir)

	participants, err := peerStore.Peers()

	if err != nil {
		return err
	}

	if participants.Len() < 2 {
		return fmt.Errorf("peers.json should define at least two peers")
	}

	b.Peers = participants

	return nil
}

func (b *Lachesis) initStore() error {
	var dbDir = fmt.Sprintf("%s/badger", b.Config.DataDir)

	if !b.Config.Store {
		b.Store = h.NewInmemStore(b.Peers, b.Config.NodeConfig.CacheSize)

		b.Config.Logger.Debug("created new in-mem store")
	} else {
		var err error

		b.Store, err = h.LoadOrCreateBadgerStore(b.Peers, b.Config.NodeConfig.CacheSize, dbDir)

		if err != nil {
			return err
		}

		if b.Store.NeedBoostrap() {
			b.Config.Logger.Debug("loaded badger store from existing database at ", dbDir)
		} else {
			b.Config.Logger.Debug("created new badger store from fresh database")
		}
	}

	return nil
}

func (b *Lachesis) initKey() error {
	if b.Config.Key == nil {
		pemKey := crypto.NewPemKey(b.Config.DataDir)

		privKey, err := pemKey.ReadKey()

		if err != nil {
			b.Config.Logger.Warn("Cannot read private key from file", err)

			privKey, err = Keygen(b.Config.DataDir)

			if err != nil {
				b.Config.Logger.Error("Cannot generate a new private key", err)

				return err
			}

			pem, _ := crypto.ToPemKey(privKey)

			b.Config.Logger.Info("Created a new key:", pem.PublicKey)
		}

		b.Config.Key = privKey
	}

	return nil
}

func (b *Lachesis) initNode() error {
	key := b.Config.Key

	nodePub := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
	n, ok := b.Peers.ByPubKey[nodePub]

	if !ok {
		return fmt.Errorf("Cannot find self pubkey in peers.json")
	}

	nodeID := n.ID

	b.Config.Logger.WithFields(logrus.Fields{
		"participants": b.Peers,
		"id":           nodeID,
	}).Debug("PARTICIPANTS")

	b.Node = node.NewNode(
		&b.Config.NodeConfig,
		nodeID,
		key,
		b.Peers,
		b.Store,
		b.Transport,
		b.Config.Proxy,
	)

	if err := b.Node.Init(); err != nil {
		return fmt.Errorf("failed to initialize node: %s", err)
	}

	return nil
}

func (b *Lachesis) initService() error {
	b.Service = service.NewService(b.Config.ServiceAddr, b.Node, b.Config.Logger)
	return nil
}

func (b *Lachesis) Init() error {
	if b.Config.Logger == nil {
		b.Config.Logger = logrus.New()
	}

	if err := b.initPeers(); err != nil {
		return err
	}

	if err := b.initStore(); err != nil {
		return err
	}

	if err := b.initTransport(); err != nil {
		return err
	}

	if err := b.initKey(); err != nil {
		return err
	}

	if err := b.initNode(); err != nil {
		return err
	}

	if err := b.initService(); err != nil {
		return err
	}

	return nil
}

func (b *Lachesis) Run() {
	go b.Service.Serve()
	b.Node.Run(true)
}

func Keygen(datadir string) (*ecdsa.PrivateKey, error) {
	pemKey := crypto.NewPemKey(datadir)

	_, err := pemKey.ReadKey()

	if err == nil {
		return nil, fmt.Errorf("Another key already lives under %s", datadir)
	}

	privKey, err := crypto.GenerateECDSAKey()

	if err != nil {
		return nil, err
	}

	if err := pemKey.WriteKey(privKey); err != nil {
		return nil, err
	}

	return privKey, nil
}
