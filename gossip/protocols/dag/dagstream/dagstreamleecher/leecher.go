package dagstreamleecher

import (
	"math/rand"
	"time"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamleecher"
	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamleecher/basepeerleecher"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream"
)

// Leecher is responsible for requesting events based on lexicographic event streams
type Leecher struct {
	*basestreamleecher.BaseLeecher

	// Callbacks
	callback Callbacks

	cfg Config

	// State
	session sessionState
	epoch   idx.Epoch

	emptyState   bool
	forceSyncing bool
	paused       bool
}

// New creates an events downloader to request events based on lexicographic event streams
func New(epoch idx.Epoch, emptyState bool, cfg Config, callback Callbacks) *Leecher {
	l := &Leecher{
		cfg:        cfg,
		callback:   callback,
		emptyState: emptyState,
		epoch:      epoch,
	}
	l.BaseLeecher = basestreamleecher.New(cfg.RecheckInterval, basestreamleecher.Callbacks{
		SelectSessionPeerCandidates: l.selectSessionPeerCandidates,
		ShouldTerminateSession:      l.shouldTerminateSession,
		StartSession:                l.startSession,
		TerminateSession:            l.terminateSession,
		OngoingSession: func() bool {
			return l.session.agent != nil
		},
		OngoingSessionPeer: func() string {
			return l.session.peer
		},
	})
	return l
}

type Callbacks struct {
	IsProcessed func(hash.Event) bool

	RequestChunk func(peer string, r dagstream.Request) error
	Suspend      func(peer string) bool
	PeerEpoch    func(peer string) idx.Epoch
}

type sessionState struct {
	agent        *basepeerleecher.BasePeerLeecher
	peer         string
	startTime    time.Time
	endTime      time.Time
	lastReceived time.Time
	try          uint32
}

func (d *Leecher) shouldTerminateSession() bool {
	if d.paused || d.session.agent.Stopped() {
		return true
	}

	noProgress := time.Since(d.session.lastReceived) >= d.cfg.BaseProgressWatchdog*time.Duration(d.session.try+5)/5
	stuck := time.Since(d.session.startTime) >= d.cfg.BaseSessionWatchdog*time.Duration(d.session.try+5)/5
	return stuck || noProgress
}

func (d *Leecher) terminateSession() {
	// force the epoch download to end
	if d.session.agent != nil {
		d.session.agent.Terminate()
		d.session.agent = nil
		d.session.endTime = time.Now()
	}
}

func (d *Leecher) Pause() {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	d.paused = true
	d.terminateSession()
}

func (d *Leecher) Resume() {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	d.paused = false
}

func (d *Leecher) selectSessionPeerCandidates() []string {
	if d.paused {
		return nil
	}
	var selected []string
	currentEpochPeers := make([]string, 0, len(d.Peers))
	futureEpochPeers := make([]string, 0, len(d.Peers))
	for p := range d.Peers {
		epoch := d.callback.PeerEpoch(p)
		if epoch == d.epoch {
			currentEpochPeers = append(currentEpochPeers, p)
		}
		if epoch > d.epoch {
			futureEpochPeers = append(futureEpochPeers, p)
		}
	}
	sinceEnd := time.Since(d.session.endTime)
	waitUntilProcessed := d.session.try == 0 || sinceEnd > d.cfg.MinSessionRestart
	hasSomethingToSync := d.session.try == 0 || len(futureEpochPeers) > 0 || sinceEnd >= d.cfg.MaxSessionRestart || d.forceSyncing
	if waitUntilProcessed && hasSomethingToSync {
		if len(futureEpochPeers) > 0 && (d.session.try%5 != 4 || len(currentEpochPeers) == 0) {
			// normally work only with peers which have a higher epoch
			selected = futureEpochPeers
		} else {
			// if above doesn't work, try peers on current epoch every 5th try
			selected = currentEpochPeers
		}
	}
	return selected
}

func getSessionID(epoch idx.Epoch, try uint32) uint32 {
	return (uint32(epoch) << 12) ^ try
}

func (d *Leecher) startSession(candidates []string) {
	peer := candidates[rand.Intn(len(candidates))]

	typ := dagstream.RequestIDs
	if d.callback.PeerEpoch(peer) > d.epoch && d.emptyState && d.session.try == 0 {
		typ = dagstream.RequestEvents
	}

	session := dagstream.Session{
		ID:    getSessionID(d.epoch, d.session.try),
		Start: d.epoch.Bytes(),
		Stop:  (d.epoch + 1).Bytes(),
	}

	d.session.agent = basepeerleecher.New(&d.Wg, d.cfg.Session, basepeerleecher.EpochDownloaderCallbacks{
		IsProcessed: func(id interface{}) bool {
			return d.callback.IsProcessed(id.(hash.Event))
		},
		RequestChunks: func(maxNum uint32, maxSize uint64, chunks uint32) error {
			return d.callback.RequestChunk(peer,
				dagstream.Request{
					Session:   session,
					Limit:     dag.Metric{Num: idx.Event(maxNum), Size: maxSize},
					Type:      typ,
					MaxChunks: chunks,
				})
		},
		Suspend: func() bool {
			return d.callback.Suspend(peer)
		},
		Done: func() bool {
			return false
		},
	})

	now := time.Now()
	d.session.startTime = now
	d.session.lastReceived = now
	d.session.endTime = now
	d.session.try++
	d.session.peer = peer

	d.session.agent.Start()

	d.forceSyncing = false
}

func (d *Leecher) OnNewEpoch(myEpoch idx.Epoch) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	if d.Terminated {
		return
	}

	d.terminateSession()

	d.epoch = myEpoch
	d.session.try = 0
	d.emptyState = true

	d.Routine()
}

func (d *Leecher) ForceSyncing() {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	d.forceSyncing = true
}

func (d *Leecher) NotifyChunkReceived(sessionID uint32, last hash.Event, done bool) error {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	if d.session.agent == nil {
		return nil
	}
	if getSessionID(d.epoch, d.session.try-1) != sessionID {
		return nil
	}

	d.session.lastReceived = time.Now()
	if done {
		d.terminateSession()
		return nil
	}
	return d.session.agent.NotifyChunkReceived(last)
}
