package epstreamleecher

import (
	"math/rand"
	"time"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamleecher"
	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamleecher/basepeerleecher"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/gossip/protocols/epochpacks/epstream"
)

// Leecher is responsible for requesting EPs based on lexicographic EPs streams
type Leecher struct {
	*basestreamleecher.BaseLeecher

	// Callbacks
	callback Callbacks

	cfg Config

	// State
	session sessionState

	forceSyncing bool
}

// New creates an EPs downloader to request EPs based on lexicographic EPs streams
func New(cfg Config, callback Callbacks) *Leecher {
	l := &Leecher{
		cfg:      cfg,
		callback: callback,
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
	LowestEpochToFetch func() idx.Epoch
	MaxEpochToFetch    func() idx.Epoch
	IsProcessed        func(idx.Epoch) bool

	RequestChunk func(peer string, r epstream.Request) error
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

	sessionID uint32

	lowestEpochToFetch idx.Epoch
}

func (d *Leecher) shouldTerminateSession() bool {
	if d.session.agent.Stopped() {
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
		if d.callback.LowestEpochToFetch() >= d.session.lowestEpochToFetch+idx.Epoch(d.cfg.Session.DefaultChunkItemsNum) {
			// reset the counter of unsuccessful sync attempts
			d.session.try = 0
		}
	}
}

func (d *Leecher) selectSessionPeerCandidates() []string {
	knowledgeablePeers := make([]string, 0, len(d.Peers))
	allPeers := make([]string, 0, len(d.Peers))
	start := d.callback.LowestEpochToFetch()
	if start >= d.callback.MaxEpochToFetch() {
		return nil
	}
	for p := range d.Peers {
		epoch := d.callback.PeerEpoch(p)
		if epoch >= start {
			knowledgeablePeers = append(knowledgeablePeers, p)
		}
		allPeers = append(allPeers, p)
	}
	sinceEnd := time.Since(d.session.endTime)
	waitUntilProcessed := d.session.try == 0 || sinceEnd > d.cfg.MinSessionRestart
	hasSomethingToSync := d.session.try == 0 || len(knowledgeablePeers) > 0 || sinceEnd >= d.cfg.MaxSessionRestart || d.forceSyncing
	if waitUntilProcessed && hasSomethingToSync {
		if len(knowledgeablePeers) > 0 && d.session.try%5 != 4 {
			// normally work only with peers which have a higher epoch
			return knowledgeablePeers
		} else {
			// if above doesn't work, try other peers on 5th try
			return allPeers
		}
	}
	return nil
}

func getSessionID(epoch idx.Epoch, try uint32) uint32 {
	return (uint32(epoch) << 12) ^ try
}

func (d *Leecher) startSession(candidates []string) {
	peer := candidates[rand.Intn(len(candidates))]

	start := d.callback.LowestEpochToFetch()
	end := d.callback.MaxEpochToFetch()
	if end <= start {
		end = start + 1
	}
	session := epstream.Session{
		ID:    getSessionID(start, d.session.try),
		Start: epstream.Locator(start),
		Stop:  epstream.Locator(end),
	}

	d.session.agent = basepeerleecher.New(&d.Wg, d.cfg.Session, basepeerleecher.EpochDownloaderCallbacks{
		IsProcessed: func(id interface{}) bool {
			lastEpoch := id.(idx.Epoch)
			return d.callback.IsProcessed(lastEpoch)
		},
		RequestChunks: func(maxNum uint32, maxSize uint64, chunks uint32) error {
			return d.callback.RequestChunk(peer,
				epstream.Request{
					Session:   session,
					Limit:     epstream.Metric{Num: idx.Epoch(maxNum), Size: maxSize},
					Type:      0,
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
	d.session.sessionID = session.ID
	d.session.lowestEpochToFetch = start

	d.session.agent.Start()

	d.forceSyncing = false
}

func (d *Leecher) ForceSyncing() {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	d.forceSyncing = true
}

func (d *Leecher) NotifyChunkReceived(sessionID uint32, lastEpoch idx.Epoch, done bool) error {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	if d.session.agent == nil {
		return nil
	}
	if d.session.sessionID != sessionID {
		return nil
	}

	d.session.lastReceived = time.Now()
	if done {
		d.terminateSession()
		return nil
	}
	return d.session.agent.NotifyChunkReceived(lastEpoch)
}
