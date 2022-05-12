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
	"github.com/Fantom-foundation/go-opera/version"
)

type VerWarcher struct {
	store *Store

	done chan struct{}
	wg   sync.WaitGroup
	logger.Instance
}

func New(store *Store) *VerWarcher {
	return &VerWarcher{
		store:    store,
		done:     make(chan struct{}),
		Instance: logger.New(),
	}
}

func (w *VerWarcher) Pause() error {
	if w.store.GetNetworkVersion() > version.AsU64() {
		return fmt.Errorf("Network upgrade %s was activated. Current node version is %s. "+
			"Please upgrade your node to continue syncing.", version.U64ToString(w.store.GetNetworkVersion()), version.AsString())
	} else if w.store.GetMissedVersion() > 0 {
		return fmt.Errorf("Node's state is dirty because node was upgraded after the network upgrade %s was activated. "+
			"Please re-sync the chain data to continue.", version.U64ToString(w.store.GetMissedVersion()))
	}
	return nil
}

func (w *VerWarcher) OnNewLog(l *types.Log) {
	if l.Address != driver.ContractAddress {
		return
	}
	if l.Topics[0] == driverpos.Topics.UpdateNetworkVersion && len(l.Data) >= 32 {
		netVersion := new(big.Int).SetBytes(l.Data[24:32]).Uint64()
		w.store.SetNetworkVersion(netVersion)
		w.log()
	}
}

func (w *VerWarcher) log() {
	if err := w.Pause(); err != nil {
		w.Log.Warn(err.Error())
	}
}

func (w *VerWarcher) Start() {
	w.log()
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(time.Second)
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
