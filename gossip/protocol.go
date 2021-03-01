package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
)

// Constants to match up protocol versions and messages
const (
	lachesis62 = 62 // derived from eth62
)

// protocolName is the official short name of the protocol used during capability negotiation.
const protocolName = "opera"

// ProtocolVersions are the supported versions of the protocol (first is primary).
var ProtocolVersions = []uint{lachesis62}

// protocolLengths are the number of implemented message corresponding to different protocol versions.
var protocolLengths = map[uint]uint64{lachesis62: EventsStreamResponse + 1}

const protocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

// protocol message codes
const (
	HandshakeMsg = 0

	// Signals about the current synchronization status.
	// The current peer's status is used during packs downloading,
	// and to estimate may peer be interested in the new event or not
	// (based on peer's epoch).
	ProgressMsg = 1

	EvmTxsMsg         = 2
	NewEvmTxHashesMsg = 3
	GetEvmTxsMsg      = 4

	// Non-aggressive events propagation. Signals about newly-connected
	// batch of events, sending only their IDs.
	NewEventIDsMsg = 5

	// Request the batch of events by IDs
	GetEventsMsg = 6
	// Contains the batch of events.
	// May be an answer to GetEventsMsg, or be sent during aggressive events propagation.
	EventsMsg = 7

	// Request a range of events by a selector
	RequestEventsStream = 8
	// Contains the requested events by RequestEventsStream
	EventsStreamResponse = 9
)

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIDMismatch
	ErrGenesisMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
	ErrEmptyMessage = 0xf00
)

func (e errCode) String() string {
	return errorToString[int(e)]
}

// XXX change once legacy code is out
var errorToString = map[int]string{
	ErrMsgTooLarge:             "Message too long",
	ErrDecode:                  "Invalid message",
	ErrInvalidMsgCode:          "Invalid message code",
	ErrProtocolVersionMismatch: "Protocol version mismatch",
	ErrNetworkIDMismatch:       "NetworkId mismatch",
	ErrGenesisMismatch:         "Genesis object mismatch",
	ErrNoStatusMsg:             "No status message",
	ErrExtraStatusMsg:          "Extra status message",
	ErrSuspendedPeer:           "Suspended peer",
	ErrEmptyMessage:            "Empty message",
}

type txPool interface {
	// AddRemotes should add the given transactions to the pool.
	AddRemotes([]*types.Transaction) []error

	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.Transactions, error)

	// SubscribeNewTxsNotify should return an event subscription of
	// NewTxsNotify and send events to the given channel.
	SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription

	Get(common.Hash) *types.Transaction

	OnlyNotExisting(hashes []common.Hash) []common.Hash
	SampleHashes(max int) []common.Hash
}

// handshakeData is the network packet for the initial handshake message
type handshakeData struct {
	ProtocolVersion uint32
	NetworkID       uint64
	Genesis         common.Hash
}

// PeerProgress is synchronization status of a peer
type PeerProgress struct {
	Epoch            idx.Epoch
	LastBlockIdx     idx.Block
	LastBlockAtropos hash.Event
	// Currently unused
	HighestLamport idx.Lamport
}

type epochChunk struct {
	SessionID uint32
	Done      bool
	IDs       hash.Events
	Events    inter.EventPayloads
}
