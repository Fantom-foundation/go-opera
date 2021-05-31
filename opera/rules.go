package opera

import (
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/genesis/evmwriter"
)

const (
	MainNetworkID   uint64 = 0xfa
	TestNetworkID   uint64 = 0xfa2
	FakeNetworkID   uint64 = 0xfa3
	DefaultEventGas uint64 = 28000
)

var DefaultVMConfig = vm.Config{
	StatePrecompiles: map[common.Address]vm.PrecompiledStateContract{
		evmwriter.ContractAddress: &evmwriter.PreCompiledContract{},
	},
}

type RulesRLP struct {
	Name      string
	NetworkID uint64

	// Graph options
	Dag DagRules

	// Epochs options
	Epochs EpochsRules

	// Blockchain options
	Blocks BlocksRules

	// Economy options
	Economy EconomyRules

	Upgrades Upgrades `rlp:"-"`
}

// Rules describes opera net.
// Note keep track of all the non-copiable variables in Copy()
type Rules RulesRLP

// GasPowerRules defines gas power rules in the consensus.
type GasPowerRules struct {
	AllocPerSec        uint64
	MaxAllocPeriod     inter.Timestamp
	StartupAllocPeriod inter.Timestamp
	MinStartupGas      uint64
}

type GasRules struct {
	MaxEventGas  uint64
	EventGas     uint64
	ParentGas    uint64
	ExtraDataGas uint64
}

type EpochsRules struct {
	MaxEpochGas      uint64
	MaxEpochDuration inter.Timestamp
}

// DagRules of Lachesis DAG (directed acyclic graph).
type DagRules struct {
	MaxParents     idx.Event
	MaxFreeParents idx.Event // maximum number of parents with no gas cost
	MaxExtraData   uint32
}

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	BlocksNum idx.Block
	Period    inter.Timestamp
}

// EconomyRules contains economy constants
type EconomyRules struct {
	BlockMissedSlack idx.Block

	Gas GasRules

	MinGasPrice *big.Int

	ShortGasPower GasPowerRules
	LongGasPower  GasPowerRules
}

// BlocksRules contains blocks constants
type BlocksRules struct {
	MaxBlockGas             uint64 // technical hard limit, gas is mostly governed by gas power allocation
	MaxEmptyBlockSkipPeriod inter.Timestamp
}

type Upgrades struct {
	Berlin bool
}

// EvmChainConfig returns ChainConfig for transactions signing and execution
func (r Rules) EvmChainConfig() *ethparams.ChainConfig {
	cfg := *ethparams.AllEthashProtocolChanges
	cfg.ChainID = new(big.Int).SetUint64(r.NetworkID)
	if !r.Upgrades.Berlin {
		cfg.BerlinBlock = nil
	}
	return &cfg
}

func MainNetRules() Rules {
	return Rules{
		Name:      "main",
		NetworkID: MainNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    DefaultEpochsRules(),
		Economy:   DefaultEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20500000,
			MaxEmptyBlockSkipPeriod: inter.Timestamp(1 * time.Minute),
		},
	}
}

func TestNetRules() Rules {
	return Rules{
		Name:      "test",
		NetworkID: TestNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    DefaultEpochsRules(),
		Economy:   DefaultEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20500000,
			MaxEmptyBlockSkipPeriod: inter.Timestamp(1 * time.Minute),
		},
	}
}

func FakeNetRules() Rules {
	return Rules{
		Name:      "fake",
		NetworkID: FakeNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    FakeNetEpochsRules(),
		Economy:   FakeEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20500000,
			MaxEmptyBlockSkipPeriod: inter.Timestamp(3 * time.Second),
		},
	}
}

// DefaultEconomyRules returns mainnet economy
func DefaultEconomyRules() EconomyRules {
	return EconomyRules{
		BlockMissedSlack: 50,
		Gas:              DefaultGasRules(),
		MinGasPrice:      big.NewInt(1e9),
		ShortGasPower:    DefaultShortGasPowerRules(),
		LongGasPower:     DefaulLongGasPowerRules(),
	}
}

// FakeEconomyRules returns fakenet economy
func FakeEconomyRules() EconomyRules {
	cfg := DefaultEconomyRules()
	cfg.ShortGasPower = FakeShortGasPowerRules()
	cfg.LongGasPower = FakeLongGasPowerRules()
	return cfg
}

func DefaultDagRules() DagRules {
	return DagRules{
		MaxParents:     10,
		MaxFreeParents: 3,
		MaxExtraData:   128,
	}
}

func DefaultEpochsRules() EpochsRules {
	return EpochsRules{
		MaxEpochGas:      1500000000,
		MaxEpochDuration: inter.Timestamp(4 * time.Hour),
	}
}

func DefaultGasRules() GasRules {
	return GasRules{
		MaxEventGas:  10000000 + DefaultEventGas,
		EventGas:     DefaultEventGas,
		ParentGas:    2400,
		ExtraDataGas: 25,
	}
}

func FakeNetEpochsRules() EpochsRules {
	cfg := DefaultEpochsRules()
	cfg.MaxEpochGas /= 5
	cfg.MaxEpochDuration = inter.Timestamp(10 * time.Minute)
	return cfg
}

// DefaulLongGasPowerRules is long-window config
func DefaulLongGasPowerRules() GasPowerRules {
	return GasPowerRules{
		AllocPerSec:        100 * DefaultEventGas,
		MaxAllocPeriod:     inter.Timestamp(60 * time.Minute),
		StartupAllocPeriod: inter.Timestamp(5 * time.Second),
		MinStartupGas:      DefaultEventGas * 20,
	}
}

// DefaultShortGasPowerRules is short-window config
func DefaultShortGasPowerRules() GasPowerRules {
	// 2x faster allocation rate, 6x lower max accumulated gas power
	cfg := DefaulLongGasPowerRules()
	cfg.AllocPerSec *= 2
	cfg.StartupAllocPeriod /= 2
	cfg.MaxAllocPeriod /= 2 * 6
	return cfg
}

// FakeLongGasPowerRules is fake long-window config
func FakeLongGasPowerRules() GasPowerRules {
	config := DefaulLongGasPowerRules()
	config.AllocPerSec *= 1000
	return config
}

// FakeShortGasPowerRules is fake short-window config
func FakeShortGasPowerRules() GasPowerRules {
	config := DefaultShortGasPowerRules()
	config.AllocPerSec *= 1000
	return config
}

func (r Rules) Copy() Rules {
	cp := r
	cp.Economy.MinGasPrice = new(big.Int).Set(r.Economy.MinGasPrice)
	return cp
}

func (r Rules) String() string {
	b, _ := json.Marshal(&r)
	return string(b)
}

// EncodeRLP is for RLP serialization.
func (r Rules) EncodeRLP(w io.Writer) error {
	// write the type
	rType := uint8(0)
	if r.Upgrades != (Upgrades{}) {
		rType = 1
		_, err := w.Write([]byte{rType})
		if err != nil {
			return err
		}
	}
	// write the main body
	rlpR := RulesRLP(r)
	err := rlp.Encode(w, &rlpR)
	if err != nil {
		return err
	}
	// write additional fields, depending on the type
	if rType > 0 {
		err := rlp.Encode(w, &r.Upgrades)
		if err != nil {
			return err
		}
	}
	return nil
}

// DecodeRLP is for RLP serialization.
func (r *Rules) DecodeRLP(s *rlp.Stream) error {
	kind, _, err := s.Kind()
	if err != nil {
		return err
	}
	// read rType
	rType := uint8(0)
	if kind == rlp.Byte {
		var b []byte
		if b, err = s.Bytes(); err != nil {
			return err
		}
		if len(b) == 0 {
			return errors.New("empty typed")
		}
		rType = b[0]
		if rType == 0 || rType > 1 {
			return errors.New("unknown type")
		}
	}
	// decode the main body
	rlpR := RulesRLP{}
	err = s.Decode(&rlpR)
	if err != nil {
		return err
	}
	*r = Rules(rlpR)
	// decode additional fields, depending on the type
	if rType >= 1 {
		err = s.Decode(&r.Upgrades)
		if err != nil {
			return err
		}
	}
	return nil
}
