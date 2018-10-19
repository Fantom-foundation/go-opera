package lachesis

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/log"
	"github.com/andrecronje/lachesis/src/net"
	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/service"
	"github.com/sirupsen/logrus"
)

type Lachesis struct {
	Config    *LachesisConfig
	Node      *node.Node
	Transport net.Transport
	Store     poset.Store
	Peers     *peers.Peers
	Service   *service.Service
}

func NewLachesis(config *LachesisConfig) *Lachesis {
	engine := &Lachesis{
		Config: config,
	}

	return engine
}

func (l *Lachesis) initTransport() error {
	transport, err := net.NewTCPTransport(
		l.Config.BindAddr,
		nil,
		l.Config.MaxPool,
		l.Config.NodeConfig.TCPTimeout,
		l.Config.Logger,
	)

	if err != nil {
		return err
	}

	l.Transport = transport

	return nil
}

func (l *Lachesis) initPeers() error {
	if !l.Config.LoadPeers {
		if l.Peers == nil {
			return fmt.Errorf("Did not load peers but none was present")
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

func (l *Lachesis) initStore() error {
	var dbDir = fmt.Sprintf("%s/badger", l.Config.DataDir)

	if !l.Config.Store {
		l.Store = poset.NewInmemStore(l.Peers, l.Config.NodeConfig.CacheSize)

		l.Config.Logger.Debug("created new in-mem store")
	} else {
		var err error

		l.Config.Logger.WithField("path", l.Config.BadgerDir()).Debug("Attempting to load or create database")
 		l.Store, err = poset.LoadOrCreateBadgerStore(l.Peers, l.Config.NodeConfig.CacheSize, dbDir)

		if err != nil {
			return err
		}

		if l.Store.NeedBoostrap() {
			l.Config.Logger.Debug("loaded badger store from existing database at ", dbDir)
		} else {
			l.Config.Logger.Debug("created new badger store from fresh database")
		}
	}

	return nil
}

func (l *Lachesis) initKey() error {
	if l.Config.Key == nil {
		pemKey := crypto.NewPemKey(l.Config.DataDir)

		privKey, err := pemKey.ReadKey()

		if err != nil {
			l.Config.Logger.Warn("Cannot read private key from file", err)

			privKey, err = Keygen(l.Config.DataDir)

			if err != nil {
				l.Config.Logger.Error("Cannot generate a new private key", err)

				return err
			}

			pem, _ := crypto.ToPemKey(privKey)

			l.Config.Logger.Info("Created a new key:", pem.PublicKey)
		}

		l.Config.Key = privKey
	}

	return nil
}

func (l *Lachesis) initNode() error {
	key := l.Config.Key

	nodePub := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
	n, ok := l.Peers.ByPubKey[nodePub]

	if !ok {
		return fmt.Errorf("Cannot find self pubkey in peers.json")
	}

	nodeID := n.ID

	l.Config.Logger.WithFields(logrus.Fields{
		"participants": l.Peers,
		"id":           nodeID,
	}).Debug("PARTICIPANTS")

	l.Node = node.NewNode(
		&l.Config.NodeConfig,
		nodeID,
		key,
		l.Peers,
		l.Store,
		l.Transport,
		l.Config.Proxy,
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

func (l *Lachesis) Run() {
	if l.Service != nil {
		go l.Service.Serve()
	}
	l.Node.Run(true)
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
