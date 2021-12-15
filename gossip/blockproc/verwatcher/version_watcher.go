package verwatcher

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver/driverpos"
	"github.com/Fantom-foundation/go-opera/utils/errlock"
	"github.com/Fantom-foundation/go-opera/version"
)

type VerWarcher struct {
	cfg   Config
	store *Store

	done chan struct{}
	wg   sync.WaitGroup
	logger.Instance
}

func New(cfg Config, store *Store) *VerWarcher {
	return &VerWarcher{
		cfg:      cfg,
		store:    store,
		done:     make(chan struct{}),
		Instance: logger.New(),
	}
}

func (w *VerWarcher) OnNewLog(l *types.Log) {
	if l.Address != driver.ContractAddress {
		return
	}
	if l.Topics[0] == driverpos.Topics.UpdateNetworkVersion && len(l.Data) >= 32 {
		netVersion := new(big.Int).SetBytes(l.Data[24:32]).Uint64()
		w.store.SetNetworkVersion(netVersion)
		if netVersion > version.AsU64() {
			if w.cfg.ShutDownIfNotUpgraded {
				errlock.Permanent(fmt.Errorf("The network's supported version of %s was activated at block %d.\n"+
					"Node's current version is %s and a shutdown is required as ShutDownIfNotUpgraded flag was set.\n"+
					"Please upgrade the node to continue.", version.U64ToString(netVersion), l.BlockNumber, version.AsString()))
				panic("unreachable")
			} else if w.store.GetMissedVersion() == 0 {
				w.store.SetMissedVersion(netVersion)
			}
		}
		w.log()
	}
}

func (w *VerWarcher) log() {
	if w.cfg.WarningIfNotUpgradedEvery == 0 {
		return
	}
	if w.store.GetNetworkVersion() > version.AsU64() {
		w.Log.Warn(fmt.Sprintf("Network upgrade %s was activated. Current node version is %s. "+
			"Please upgrade your node and re-sync the chain data.", version.U64ToString(w.store.GetNetworkVersion()), version.AsString()))
	} else if w.store.GetMissedVersion() > 0 {
		w.Log.Warn(fmt.Sprintf("Node's state is dirty because node was upgraded after the network upgrade %s was activated. "+
			"Please re-sync the chain data to continue.", version.U64ToString(w.store.GetMissedVersion())))
	}
}

func (w *VerWarcher) Start() {
	if w.cfg.WarningIfNotUpgradedEvery == 0 {
		return
	}
	w.log()
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.cfg.WarningIfNotUpgradedEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.log()
			case <-w.done:
				return
			}
		}
	}()
}

func (w *VerWarcher) Stop() {
	close(w.done)
	w.wg.Wait()
}
