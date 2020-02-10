package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// Constants to match up protocol versions and messages
const (
	lachesis62 = 62 // derived from eth62
)

// protocolName is the official short name of the protocol used during capability negotiation.
const protocolName = "lachesis"

// ProtocolVersions are the supported versions of the protocol (first is primary).
var ProtocolVersions = []uint{lachesis62}

// protocolLengths are the number of implemented message corresponding to different protocol versions.
var protocolLengths = map[uint]uint64{lachesis62: PackMsg + 1}

const protocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

// protocol message codes
const (
	// Protocol messages belonging to eth/62
	EthStatusMsg = 0x00
	EvmTxMsg     = 0x02

	// Protocol messages belonging to lachesis/62

	// Signals about the current synchronization status.
	// The current peer's status is used during packs downloading,
	// and to estimate may peer be interested in the new event or not
	// (based on peer's epoch).
	ProgressMsg = 0xf0

	// Non-aggressive events propagation. Signals about newly-connected
	// batch of events, sending only their IDs.
	NewEventHashesMsg = 0xf1

	// Request the batch of events by IDs
	GetEventsMsg = 0xf2
	// Contains the batch of events.
	// May be an answer to GetEventsMsg, or be sent during aggressive events propagation.
	EventsMsg = 0xf3

	// Request pack infos by epoch:pack indexes
	GetPackInfosMsg = 0xf4
	// Contains the requested pack infos. An answer to GetPackInfosMsg.
	PackInfosMsg = 0xf5

	// Request pack by epoch:pack index
	GetPackMsg = 0xf6
	// Contains the requested pack. An answer to GetPackMsg.
	PackMsg = 0xf7
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
}

// ethStatusData is the network packet for the status message. It's used for compatibility with some ETH wallets.
type ethStatusData struct {
	ProtocolVersion   uint32
	NetworkID         uint64
	DummyTD           *big.Int
	DummyCurrentBlock common.Hash
	Genesis           common.Hash
}

// PeerProgress is synchronization status of a peer
type PeerProgress struct {
	Epoch        idx.Epoch
	NumOfBlocks  idx.Block
	LastPackInfo PackInfo
	LastBlock    hash.Event
}

type packInfosData struct {
	Epoch           idx.Epoch
	TotalNumOfPacks idx.Pack // in specified epoch
	Infos           []PackInfo
}

type packInfosDataRLP struct {
	Epoch           idx.Epoch
	TotalNumOfPacks idx.Pack // in specified epoch
	RawInfos        []rlp.RawValue
}

type getPackInfosData struct {
	Epoch   idx.Epoch
	Indexes []idx.Pack
}

type getPackData struct {
	Epoch idx.Epoch
	Index idx.Pack
}

type packData struct {
	Epoch idx.Epoch
	Index idx.Pack
	IDs   hash.Events
}
