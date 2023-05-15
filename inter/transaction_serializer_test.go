package inter

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/go-opera/utils/cser"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestSerializeLegacyTx(t *testing.T) {
	require := require.New(t)

	addr := common.HexToAddress("727fc6a68321b754475c668a6abfb6e9e71c169a")
	value := big.NewInt(10)
	value = value.Mul(value, big.NewInt(1_000_000_000))
	value = value.Mul(value, big.NewInt(1_000_000_000))

	data, _ := hex.DecodeString("a9059cbb000000000213ed0f886efd100b67c7e4ec0a85a7d20dc971600000000000000000000015af1d78b58c4000")

	r := new(big.Int)
	r.SetString("be67e0a07db67da8d446f76add590e54b6e92cb6b8f9835aeb67540579a27717", 16)

	s := new(big.Int)
	s.SetString("2d690516512020171c1ec870f6ff45398cc8609250326be89915fb538e7bd718", 16)

	v := types.NewTx(&types.LegacyTx{
		Nonce:    12,
		GasPrice: big.NewInt(20_000_000_000),
		Gas:      21_000,
		To:       &addr,
		Value:    value,
		Data:     data,
		V:        big.NewInt(40),
		R:        r,
		S:        s,
	})

	encoded, err := cser.MarshalBinaryAdapter(func(w *cser.Writer) error {
		return TransactionMarshalCSER(w, v)
	})
	require.NoError(err)
	require.Equal("0c08520504a817c800088ac7230489e80000727fc6a68321b754475c668a6abfb6e9e71c169a2fa9059cbb000000000213ed0f886efd100b67c7e4ec0a85a7d20dc971600000000000000000000015af1d78b58c40000128be67e0a07db67da8d446f76add590e54b6e92cb6b8f9835aeb67540579a277172d690516512020171c1ec870f6ff45398cc8609250326be89915fb538e7bd71848320183", hex.EncodeToString(encoded))
}

func TestSerializeAccessListTx(t *testing.T) {
	require := require.New(t)

	addr := common.HexToAddress("811a752c8cd697e3cb27279c330ed1ada745a8d7")
	value := big.NewInt(2)
	value = value.Mul(value, big.NewInt(1_000_000_000))
	value = value.Mul(value, big.NewInt(1_000_000_000))

	data, _ := hex.DecodeString("6ebaf477f83e051589c1188bcc6ddccd")

	r := new(big.Int)
	r.SetString("36b241b061a36a32ab7fe86c7aa9eb592dd59018cd0443adc0903590c16b02b0", 16)

	s := new(big.Int)
	s.SetString("5edcc541b4741c5cc6dd347c5ed9577ef293a62787b4510465fadbfe39ee4094", 16)

	v := types.NewTx(&types.AccessListTx{
		ChainID:  big.NewInt(5),
		Nonce:    7,
		GasPrice: big.NewInt(30_000_000_000),
		Gas:      5_748_100,
		To:       &addr,
		Value:    value,
		Data:     data,
		AccessList: []types.AccessTuple{
			types.AccessTuple{
				Address: common.HexToAddress("de0b295669a9fd93d5f28d9ec85e40f4cb697bae"),
				StorageKeys: []common.Hash{
					common.HexToHash("0000000000000000000000000000000000000000000000000000000000000003"),
					common.HexToHash("0000000000000000000000000000000000000000000000000000000000000007"),
				},
			},
			types.AccessTuple{
				Address: common.HexToAddress("bb9bc244d798123fde783fcc1c72d3bb8c189413"),
			},
		},
		V: big.NewInt(45),
		R: r,
		S: s,
	})

	encoded, err := cser.MarshalBinaryAdapter(func(w *cser.Writer) error {
		return TransactionMarshalCSER(w, v)
	})
	require.NoError(err)
	require.Equal("010784b5570506fc23ac00081bc16d674ec80000811a752c8cd697e3cb27279c330ed1ada745a8d7106ebaf477f83e051589c1188bcc6ddccd012d36b241b061a36a32ab7fe86c7aa9eb592dd59018cd0443adc0903590c16b02b05edcc541b4741c5cc6dd347c5ed9577ef293a62787b4510465fadbfe39ee4094010502de0b295669a9fd93d5f28d9ec85e40f4cb697bae0200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000007bb9bc244d798123fde783fcc1c72d3bb8c1894130000944c020085", hex.EncodeToString(encoded))
}

func TestSerializeDynamicFeeTx(t *testing.T) {
	require := require.New(t)

	addr := common.HexToAddress("811a752c8cd697e3cb27279c330ed1ada745a8d7")
	value := big.NewInt(2)
	value = value.Mul(value, big.NewInt(1_000_000_000))
	value = value.Mul(value, big.NewInt(1_000_000_000))

	data, _ := hex.DecodeString("6ebaf477f83e051589c1188bcc6ddccd")

	r := new(big.Int)
	r.SetString("36b241b061a36a32ab7fe86c7aa9eb592dd59018cd0443adc0903590c16b02b0", 16)

	s := new(big.Int)
	s.SetString("5edcc541b4741c5cc6dd347c5ed9577ef293a62787b4510465fadbfe39ee4094", 16)

	v := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(5),
		Nonce:     7,
		GasTipCap: big.NewInt(10_000_000_000),
		GasFeeCap: big.NewInt(30_000_000_000),
		Gas:       5_748_100,
		To:        &addr,
		Value:     value,
		Data:      data,
		AccessList: []types.AccessTuple{
			types.AccessTuple{
				Address: common.HexToAddress("de0b295669a9fd93d5f28d9ec85e40f4cb697bae"),
				StorageKeys: []common.Hash{
					common.HexToHash("0000000000000000000000000000000000000000000000000000000000000003"),
					common.HexToHash("0000000000000000000000000000000000000000000000000000000000000007"),
				},
			},
			types.AccessTuple{
				Address: common.HexToAddress("bb9bc244d798123fde783fcc1c72d3bb8c189413"),
			},
		},
		V: big.NewInt(45),
		R: r,
		S: s,
	})

	encoded, err := cser.MarshalBinaryAdapter(func(w *cser.Writer) error {
		return TransactionMarshalCSER(w, v)
	})
	require.NoError(err)
	require.Equal("020784b5570502540be4000506fc23ac00081bc16d674ec80000811a752c8cd697e3cb27279c330ed1ada745a8d7106ebaf477f83e051589c1188bcc6ddccd012d36b241b061a36a32ab7fe86c7aa9eb592dd59018cd0443adc0903590c16b02b05edcc541b4741c5cc6dd347c5ed9577ef293a62787b4510465fadbfe39ee4094010502de0b295669a9fd93d5f28d9ec85e40f4cb697bae0200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000007bb9bc244d798123fde783fcc1c72d3bb8c18941300009464120085", hex.EncodeToString(encoded))
}
