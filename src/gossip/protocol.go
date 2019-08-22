package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
)

// Constants to match up protocol versions and messages
const (
	fantom62 = 62 // derived from eth62
)

// protocolName is the official short name of the protocol used during capability negotiation.
const protocolName = "fantom"

// ProtocolVersions are the supported versions of the protocol (first is primary).
var ProtocolVersions = []uint{fantom62}

// protocolLengths are the number of implemented message corresponding to different protocol versions.
var protocolLengths = map[uint]uint64{fantom62: PackMsg + 1}

const protocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

// protocol message codes
const (
	// Protocol messages belonging to eth/62
	EthStatusMsg = 0x00
	EvmTxMsg     = 0x02

	// Protocol messages belonging to fantom/62

	ProgressMsg = 0xf0

	NewEventHashesMsg = 0xf1

	GetEventsMsg = 0xf2
	EventsMsg    = 0xf3

	GetPackInfosMsg = 0xf4
	PackInfosMsg    = 0xf5

	GetPackMsg = 0xf6
	PackMsg    = 0xf7
)

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
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
	ErrNetworkIdMismatch:       "NetworkId mismatch",
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

	// SubscribeNewTxsEvent should return an event subscription of
	// NewTxsEvent and send events to the given channel.
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription
}

// ethStatusData is the network packet for the status message. It's used for compatibility with some ETH wallets.
type ethStatusData struct {
	ProtocolVersion   uint32
	NetworkId         uint64
	DummyTD           *big.Int
	DummyCurrentBlock hash.Hash
	Genesis           hash.Hash
}

type PeerProgress struct {
	Epoch        idx.SuperFrame
	NumOfBlocks  idx.Block
	LastPackInfo PackInfo
	LastBlock    hash.Event
}

type packInfosData struct {
	Epoch           idx.SuperFrame
	TotalNumOfPacks idx.Pack // in specified epoch
	Infos           []PackInfo
}

type packInfosDataRLP struct {
	Epoch           idx.SuperFrame
	TotalNumOfPacks idx.Pack // in specified epoch
	RawInfos        []rlp.RawValue
}

type getPackInfosData struct {
	Epoch   idx.SuperFrame
	Indexes []idx.Pack
}

type getPackData struct {
	Epoch idx.SuperFrame
	Index idx.Pack
}

type packData struct {
	Epoch idx.SuperFrame
	Index idx.Pack
	Ids   hash.Events
}
