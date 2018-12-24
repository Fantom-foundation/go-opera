package node

import (
	"crypto/ecdsa"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/log"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

const (
	// TODO: collect the similar magic constants in protocol config
	// MaxReceiveMessageSize is size limitation of txs in bytes
	MaxEventsPayloadSize = 100 * 1024 * 1024
)

var (
	ErrTooBigTx = fmt.Errorf("Transaction too big")
)

// Core struct that controls the consensus, transaction, and communication
type Core struct {
	id     int64
	key    *ecdsa.PrivateKey
	pubKey []byte
	hexID  string
	poset  *poset.Poset

	inDegrees map[string]uint64

	participants *peers.Peers // [PubKey] => id
	head         string
	Seq          int64

	transactionPool         [][]byte
	internalTransactionPool []poset.InternalTransaction
	blockSignaturePool      []poset.BlockSignature

	logger *logrus.Entry

	addSelfEventBlockLocker       sync.Mutex
	transactionPoolLocker         sync.RWMutex
	internalTransactionPoolLocker sync.RWMutex
	blockSignaturePoolLocker      sync.RWMutex
}

// NewCore creates a new core struct
func NewCore(id int64, key *ecdsa.PrivateKey, participants *peers.Peers,
	store poset.Store, commitCh chan poset.Block, logger *logrus.Logger) *Core {

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
		lachesis_log.NewLocal(logger, logger.Level.String())
	}
	logEntry := logger.WithField("id", id)

	inDegrees := make(map[string]uint64)
	for pubKey := range participants.ByPubKey {
		inDegrees[pubKey] = 0
	}

	p2 := poset.NewPoset(participants, store, commitCh, logEntry)
	core := &Core{
		id:                      id,
		key:                     key,
		poset:                   p2,
		inDegrees:               inDegrees,
		participants:            participants,
		transactionPool:         [][]byte{},
		internalTransactionPool: []poset.InternalTransaction{},
		blockSignaturePool:      []poset.BlockSignature{},
		logger:                  logEntry,
		head:                    "",
		Seq:                     -1,
	}

	p2.SetCore(core)

	return core
}

// ID returns the ID of this core
func (c *Core) ID() int64 {
	return c.id
}

// PubKey returns the public key of this core
func (c *Core) PubKey() []byte {
	if c.pubKey == nil {
		c.pubKey = crypto.FromECDSAPub(&c.key.PublicKey)
	}
	return c.pubKey
}

// HexID returns the Hex representation of the public key
func (c *Core) HexID() string {
	if c.hexID == "" {
		pubKey := c.PubKey()
		c.hexID = fmt.Sprintf("0x%X", pubKey)
	}
	return c.hexID
}

// Head returns the current chain head for this core
func (c *Core) Head() string {
	return c.head
}

// Heights returns map with heights for each participants
func (c *Core) Heights() map[string]uint64 {
	heights := make(map[string]uint64)
	for pubKey := range c.participants.ByPubKey {
		participantEvents, err := c.poset.Store.ParticipantEvents(pubKey, -1)
		if err == nil {
			heights[pubKey] = uint64(len(participantEvents))
		} else {
			heights[pubKey] = 0
		}
	}
	return heights
}

// InDegrees returns all vertexes from other nodes that reference this top event block
func (c *Core) InDegrees() map[string]uint64 {
	return c.inDegrees
}

// SetHeadAndSeq calculates and sets the current head for the chain
func (c *Core) SetHeadAndSeq() error {

	var head string
	var seq int64

	last, isRoot, err := c.poset.Store.LastEventFrom(c.HexID())
	if err != nil {
		return err
	}

	if isRoot {
		root, err := c.poset.Store.GetRoot(c.HexID())
		if err != nil {
			return err
		}
		head = root.SelfParent.Hash
		seq = root.SelfParent.Index
	} else {
		lastEvent, err := c.GetEventBlock(last)
		if err != nil {
			return err
		}
		head = last
		seq = lastEvent.Index()
	}

	c.head = head
	c.Seq = seq

	c.logger.WithFields(logrus.Fields{
		"core.head": c.head,
		"core.Seq":  c.Seq,
		"is_root":   isRoot,
	}).Debugf("SetHeadAndSeq()")

	return nil
}

// Bootstrap the poset with default values
func (c *Core) Bootstrap() error {
	if err := c.poset.Bootstrap(); err != nil {
		return err
	}
	c.bootstrapInDegrees()
	return nil
}

func (c *Core) bootstrapInDegrees() {
	for pubKey := range c.participants.ByPubKey {
		c.inDegrees[pubKey] = 0
		eventHash, _, err := c.poset.Store.LastEventFrom(pubKey)
		if err != nil {
			continue
		}
		for otherPubKey := range c.participants.ByPubKey {
			if otherPubKey == pubKey {
				continue
			}
			events, err := c.poset.Store.ParticipantEvents(otherPubKey, -1)
			if err != nil {
				continue
			}
			for _, eh := range events {
				event, err := c.poset.Store.GetEventBlock(eh)
				if err != nil {
					continue
				}
				if event.OtherParent() == eventHash {
					c.inDegrees[pubKey]++
				}
			}
		}
	}
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// SignAndInsertSelfEvent signs and inserts a self generated event block
func (c *Core) SignAndInsertSelfEvent(event poset.Event) error {
	if err := c.poset.SetWireInfoAndSign(&event, c.key); err != nil {
		return err
	}

	return c.InsertEvent(event, true)
}

// InsertEvent inserts an unknown event block
func (c *Core) InsertEvent(event poset.Event, setWireInfo bool) error {

	c.logger.WithFields(logrus.Fields{
		"event":      event,
		"creator":    event.GetCreator(),
		"selfParent": event.SelfParent(),
		"index":      event.Index(),
		"hex":        event.Hex(),
	}).Debug("InsertEvent(event poset.Event, setWireInfo bool)")

	if err := c.poset.InsertEvent(event, setWireInfo); err != nil {
		return err
	}

	if event.GetCreator() == c.HexID() {
		c.head = event.Hex()
		c.Seq = event.Index()
	}

	c.inDegrees[event.GetCreator()] = 0

	if otherEvent, err := c.poset.Store.GetEventBlock(event.OtherParent()); err == nil {
		c.inDegrees[otherEvent.GetCreator()]++
	}
	return nil
}

// KnownEvents returns all known event blocks
func (c *Core) KnownEvents() map[int64]int64 {
	return c.poset.Store.KnownEvents()
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// SignBlock sign a block to register it as an anchor block
func (c *Core) SignBlock(block poset.Block) (poset.BlockSignature, error) {
	sig, err := block.Sign(c.key)
	if err != nil {
		return poset.BlockSignature{}, err
	}
	if err := block.SetSignature(sig); err != nil {
		return poset.BlockSignature{}, err
	}
	return sig, c.poset.Store.SetBlock(block)
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// OverSyncLimit checks if the unknown events is over the sync limit and if the node should catch up
func (c *Core) OverSyncLimit(knownEvents map[int64]int64, syncLimit int64) bool {
	totUnknown := int64(0)
	myKnownEvents := c.KnownEvents()
	for i, li := range myKnownEvents {
		if li > knownEvents[i] {
			totUnknown += li - knownEvents[i]
		}
	}
	return totUnknown > syncLimit
}

// GetAnchorBlockWithFrame returns the current anchor block and their frame
func (c *Core) GetAnchorBlockWithFrame() (poset.Block, poset.Frame, error) {
	return c.poset.GetAnchorBlockWithFrame()
}

// EventDiff returns events that c knows about and are not in 'known'
func (c *Core) EventDiff(known map[int64]int64) (events []poset.Event, err error) {
	var unknown []poset.Event
	// known represents the index of the last event known for every participant
	// compare this to our view of events and fill unknown with events that we know of
	// and the other doesn't
	for id, ct := range known {
		peer := c.participants.ByID[id]
		if peer == nil {
			// unknown peer detected.
			// TODO: we should handle this nicely
			continue
		}
		// get participant Events with index > ct
		participantEvents, err := c.poset.Store.ParticipantEvents(peer.PubKeyHex, ct)
		if err != nil {
			return []poset.Event{}, err
		}
		for _, e := range participantEvents {
			ev, err := c.poset.Store.GetEventBlock(e)
			if err != nil {
				return []poset.Event{}, err
			}
			c.logger.WithFields(logrus.Fields{
				"event":      ev,
				"creator":    ev.GetCreator(),
				"selfParent": ev.SelfParent(),
				"index":      ev.Index(),
				"hex":        ev.Hex(),
			}).Debugf("Sending Unknown Event")
			unknown = append(unknown, ev)
		}
	}
	sort.Stable(poset.ByTopologicalOrder(unknown))

	return unknown, nil
}

// Sync unknown events into our poset
func (c *Core) Sync(unknownEvents []poset.WireEvent) error {

	c.logger.WithFields(logrus.Fields{
		"unknown_events":              len(unknownEvents),
		"transaction_pool":            c.GetTransactionPoolCount(),
		"internal_transaction_pool":   c.GetInternalTransactionPoolCount(),
		"block_signature_pool":        c.GetBlockSignaturePoolCount(),
		"c.poset.PendingLoadedEvents": c.poset.GetPendingLoadedEvents(),
	}).Debug("Sync(unknownEventBlocks []poset.EventBlock)")

	myKnownEvents := c.KnownEvents()
	otherHead := ""
	// add unknown events
	for k, we := range unknownEvents {
		c.logger.WithFields(logrus.Fields{
			"unknown_events": we,
		}).Debug("unknownEvents")
		ev, err := c.poset.ReadWireInfo(we)
		if err != nil {
			c.logger.WithField("EventBlock", we).Errorf("c.poset.ReadEventBlockInfo(we)")
			return err

		}
		if ev.Index() > myKnownEvents[ev.CreatorID()] {
			ev.SetLamportTimestamp(poset.LamportTimestampNIL)
			ev.SetRound(poset.RoundNIL)
			ev.SetRoundReceived(poset.RoundNIL)
			if err := c.InsertEvent(*ev, false); err != nil {
				c.logger.Error("SYNC: INSERT ERR", err)
				return err
			}
		}

		// assume last event corresponds to other-head
		if k == len(unknownEvents)-1 {
			otherHead = ev.Hex()
		}
	}

	// create new event with self head and other head only if there are pending
	// loaded events or the pools are not empty
	if c.poset.GetPendingLoadedEvents() > 0 ||
		c.GetTransactionPoolCount() > 0 ||
		c.GetInternalTransactionPoolCount() > 0 ||
		c.GetBlockSignaturePoolCount() > 0 {
		return c.AddSelfEventBlock(otherHead)
	}
	return nil
}

// FastForward catch up to another peer if too far behind
func (c *Core) FastForward(peer string, block poset.Block, frame poset.Frame) error {

	// Check Block Signatures
	err := c.poset.CheckBlock(block)
	if err != nil {
		return err
	}

	// Check Frame Hash
	frameHash, err := frame.Hash()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(block.GetFrameHash(), frameHash) {
		return fmt.Errorf("invalid Frame Hash")
	}

	err = c.poset.Reset(block, frame)
	if err != nil {
		return err
	}

	err = c.SetHeadAndSeq()
	if err != nil {
		return err
	}

	err = c.RunConsensus()
	if err != nil {
		return err
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AddSelfEventBlock adds an event block created by this node
func (c *Core) AddSelfEventBlock(otherHead string) error {

	c.addSelfEventBlockLocker.Lock()
	defer c.addSelfEventBlockLocker.Unlock()

	// Get flag tables from parents
	parentEvent, errSelf := c.poset.Store.GetEventBlock(c.head)
	if errSelf != nil {
		c.logger.Warnf("failed to get parent: %s", errSelf)
	}
	otherParentEvent, errOther := c.poset.Store.GetEventBlock(otherHead)
	if errOther != nil {
		c.logger.Warnf("failed to get other parent: %s", errOther)
	}

	var (
		flagTable map[string]int64
		err       error
	)

	if errSelf != nil {
		flagTable = map[string]int64{c.head: 1}
	} else {
		flagTable, err = parentEvent.GetFlagTable()
		if err != nil {
			return fmt.Errorf("failed to get self flag table: %s", err)
		}
	}

	if errOther == nil {
		flagTable, err = otherParentEvent.MergeFlagTable(flagTable)
		if err != nil {
			return fmt.Errorf("failed to marge flag tables: %s", err)
		}
	}

	// get transactions batch for new Event
	c.transactionPoolLocker.Lock()
	var payloadSize, nTxs int
	for nTxs = 0; nTxs < len(c.transactionPool); nTxs++ {
		// NOTE: if len(tx)>MaxEventsPayloadSize it will be payloadSize>MaxEventsPayloadSize
		txSize := len(c.transactionPool[nTxs])
		if nTxs > 0 && payloadSize >= (MaxEventsPayloadSize-txSize) {
			break
		}
		payloadSize += txSize
	}
	batch := c.transactionPool[0:nTxs]
	c.transactionPool = c.transactionPool[nTxs:]
	c.transactionPoolLocker.Unlock()

	// create new event with self head and empty other parent
	newHead := poset.NewEvent(batch,
		c.internalTransactionPool,
		c.blockSignaturePool,
		[]string{c.head, otherHead}, c.PubKey(), c.Seq+1, flagTable)

	if err := c.SignAndInsertSelfEvent(newHead); err != nil {
		// put batch back to transactionPool
		c.transactionPoolLocker.Lock()
		c.transactionPool = append(batch, c.transactionPool...)
		c.transactionPoolLocker.Unlock()
		return fmt.Errorf("newHead := poset.NewEventBlock: %s", err)
	}
	c.logger.WithFields(logrus.Fields{
		"transactions":          nTxs,
		"internal_transactions": c.GetInternalTransactionPoolCount(),
		"block_signatures":      c.GetBlockSignaturePoolCount(),
	}).Debug("newHead := poset.NewEventBlock")

	c.internalTransactionPoolLocker.Lock()
	c.internalTransactionPool = []poset.InternalTransaction{}
	c.internalTransactionPoolLocker.Unlock()

	// retain c.blockSignaturePool until c.transactionPool is empty
	// FIXIT: is there any better strategy?
	if c.GetTransactionPoolCount() == 0 {
		c.blockSignaturePoolLocker.Lock()
		c.blockSignaturePool = []poset.BlockSignature{}
		c.blockSignaturePoolLocker.Unlock()
	}

	return nil
}

// FromWire converts wire events into event blocks (that were transported)
func (c *Core) FromWire(wireEvents []poset.WireEvent) ([]poset.Event, error) {
	events := make([]poset.Event, len(wireEvents))
	for i, w := range wireEvents {
		ev, err := c.poset.ReadWireInfo(w)
		if err != nil {
			return nil, err
		}
		events[i] = *ev
	}
	return events, nil
}

// ToWire converts event blocks into wire events (to be transported)
func (c *Core) ToWire(events []poset.Event) ([]poset.WireEvent, error) {
	wireEvents := make([]poset.WireEvent, len(events))
	for i, e := range events {
		wireEvents[i] = e.ToWire()
	}
	return wireEvents, nil
}

// RunConsensus is the core consensus mechanism, this checks rounds / frames and creates blocks
func (c *Core) RunConsensus() error {
	start := time.Now()
	err := c.poset.DivideRounds()
	c.logger.WithField("Duration", time.Since(start).Nanoseconds()).Debug("c.poset.DivideAtropos()")
	if err != nil {
		c.logger.WithField("Error", err).Error("c.poset.DivideAtropos()")
		return err
	}

	start = time.Now()
	err = c.poset.DecideAtropos()
	c.logger.WithField("Duration", time.Since(start).Nanoseconds()).Debug("c.poset.DecideClotho()")
	if err != nil {
		c.logger.WithField("Error", err).Error("c.poset.DecideClotho()")
		return err
	}

	start = time.Now()
	err = c.poset.DecideRoundReceived()
	c.logger.WithField("Duration", time.Since(start).Nanoseconds()).Debug("c.poset.DecideAtroposRoundReceived()")
	if err != nil {
		c.logger.WithField("Error", err).Error("c.poset.DecideAtroposRoundReceived()")
		return err
	}

	start = time.Now()
	err = c.poset.ProcessDecidedRounds()
	c.logger.WithField("Duration", time.Since(start).Nanoseconds()).Debug("c.poset.ProcessAtroposRounds()")
	if err != nil {
		c.logger.WithField("Error", err).Error("c.poset.ProcessAtroposRounds()")
		return err
	}

	start = time.Now()
	err = c.poset.ProcessSigPool()
	c.logger.WithField("Duration", time.Since(start).Nanoseconds()).Debug("c.poset.ProcessSigPool()")
	if err != nil {
		c.logger.WithField("Error", err).Error("c.poset.ProcessSigPool()")
		return err
	}

	c.logger.WithFields(logrus.Fields{
		"transaction_pool":            c.GetTransactionPoolCount(),
		"block_signature_pool":        c.GetBlockSignaturePoolCount(),
		"c.poset.pendingLoadedEvents": c.poset.GetPendingLoadedEvents(),
	}).Debug("c.RunConsensus()")

	return nil
}

// AddTransactions add transactions to the pending pool
func (c *Core) AddTransactions(txs [][]byte) error {
	for _, tx := range txs {
		if len(tx) > MaxEventsPayloadSize {
			return ErrTooBigTx
		}
	}
	c.transactionPoolLocker.Lock()
	defer c.transactionPoolLocker.Unlock()
	c.transactionPool = append(c.transactionPool, txs...)
	return nil
}

// AddInternalTransactions add internal transactions to the pending pool
func (c *Core) AddInternalTransactions(txs []poset.InternalTransaction) {
	c.internalTransactionPoolLocker.Lock()
	defer c.internalTransactionPoolLocker.Unlock()
	c.internalTransactionPool = append(c.internalTransactionPool, txs...)
}

// AddBlockSignature add block signatures to the pending pool
func (c *Core) AddBlockSignature(bs poset.BlockSignature) {
	c.blockSignaturePoolLocker.Lock()
	defer c.blockSignaturePoolLocker.Unlock()
	c.blockSignaturePool = append(c.blockSignaturePool, bs)
}

// GetHead get the current latest event block head
func (c *Core) GetHead() (poset.Event, error) {
	return c.poset.Store.GetEventBlock(c.head)
}

// GetEventBlock get a specific event block for the hash provided
func (c *Core) GetEventBlock(hash string) (poset.Event, error) {
	return c.poset.Store.GetEventBlock(hash)
}

// GetEventBlockTransactions get all transactions in an event block
func (c *Core) GetEventBlockTransactions(hash string) ([][]byte, error) {
	var txs [][]byte
	ex, err := c.GetEventBlock(hash)
	if err != nil {
		return txs, err
	}
	txs = ex.Transactions()
	return txs, nil
}

// GetConsensusEvents get all known consensus events
func (c *Core) GetConsensusEvents() []string {
	return c.poset.Store.ConsensusEvents()
}

// GetConsensusEventsCount get the count of all known consensus events
func (c *Core) GetConsensusEventsCount() int64 {
	return c.poset.Store.ConsensusEventsCount()
}

// GetUndeterminedEvents get all unconfirmed consensus events (pending)
func (c *Core) GetUndeterminedEvents() []string {
	return c.poset.GetUndeterminedEvents()
}

// GetPendingLoadedEvents returns all pending (but already stored) events
func (c *Core) GetPendingLoadedEvents() int64 {
	return c.poset.GetPendingLoadedEvents()
}

// GetConsensusTransactions return all transactions that have reached finality
func (c *Core) GetConsensusTransactions() ([][]byte, error) {
	var txs [][]byte
	for _, e := range c.GetConsensusEvents() {
		eTxs, err := c.GetEventBlockTransactions(e)
		if err != nil {
			return txs, fmt.Errorf("GetConsensusTransactions(): %s", e)
		}
		txs = append(txs, eTxs...)
	}
	return txs, nil
}

// GetLastConsensusRound returns the last consensus round known
func (c *Core) GetLastConsensusRound() int64 {
	return c.poset.GetLastConsensusRound()
}

// GetConsensusTransactionsCount returns the count of transactions that are final
func (c *Core) GetConsensusTransactionsCount() uint64 {
	return c.poset.GetConsensusTransactionsCount()
}

// GetLastCommittedRoundEventsCount count of events in last round
func (c *Core) GetLastCommittedRoundEventsCount() int {
	return c.poset.LastCommitedRoundEvents
}

// GetLastBlockIndex retuns the latest block index
func (c *Core) GetLastBlockIndex() int64 {
	return c.poset.Store.LastBlockIndex()
}

// GetTransactionPoolCount returns the count of all pending transactions
func (c *Core) GetTransactionPoolCount() int64 {
	c.transactionPoolLocker.RLock()
	defer c.transactionPoolLocker.RUnlock()
	return int64(len(c.transactionPool))
}

// GetInternalTransactionPoolCount returns the count of all pending internal transactions
func (c *Core) GetInternalTransactionPoolCount() int64 {
	c.internalTransactionPoolLocker.RLock()
	defer c.internalTransactionPoolLocker.RUnlock()
	return int64(len(c.internalTransactionPool))
}

// GetBlockSignaturePoolCount returns the count of all pending block signatures
func (c *Core) GetBlockSignaturePoolCount() int64 {
	c.blockSignaturePoolLocker.RLock()
	defer c.blockSignaturePoolLocker.RUnlock()
	return int64(len(c.blockSignaturePool))
}
