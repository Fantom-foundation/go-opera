package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/iep"
)

// Constants to match up protocol versions and messages
const (
	FTM62           = 62
	FTM63           = 63
	ProtocolVersion = FTM63
)

// ProtocolName is the official short name of the protocol used during capability negotiation.
const ProtocolName = "opera"

// ProtocolVersions are the supported versions of the protocol (first is primary).
var ProtocolVersions = []uint{FTM62, FTM63}

// protocolLengths are the number of implemented message corresponding to different protocol versions.
var protocolLengths = map[uint]uint64{FTM62: EventsStreamResponse + 1, FTM63: EPsStreamResponse + 1}

const protocolMaxMsgSize = inter.ProtocolMaxMsgSize // Maximum cap on the size of a protocol message

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

	RequestBVsStream  = 10
	BVsStreamResponse = 11
	RequestBRsStream  = 12
	BRsStreamResponse = 13
	RequestEPsStream  = 14
	EPsStreamResponse = 15
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

type TxPool interface {
	emitter.TxPool
	// AddRemotes should add the given transactions to the pool.
	AddRemotes([]*types.Transaction) []error
	AddLocals(txs []*types.Transaction) []error
	AddLocal(tx *types.Transaction) error

	Get(common.Hash) *types.Transaction

	OnlyNotExisting(hashes []common.Hash) []common.Hash
	SampleHashes(max int) []common.Hash

	Nonce(addr common.Address) uint64
	Stats() (int, int)
	Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	ContentFrom(addr common.Address) (types.Transactions, types.Transactions)
	PendingSlice() types.Transactions
}

// HandshakeData is the network packet for the initial handshake message
type HandshakeData struct {
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

type dagChunk struct {
	SessionID uint32
	Done      bool
	IDs       hash.Events
	Events    inter.EventPayloads
}

type bvsChunk struct {
	SessionID uint32
	Done      bool
	BVs       []inter.LlrSignedBlockVotes
}

type brsChunk struct {
	SessionID uint32
	Done      bool
	BRs       []ibr.LlrIdxFullBlockRecord
}

type epsChunk struct {
	SessionID uint32
	Done      bool
	EPs       []iep.LlrEpochPack
}
