package inter

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/utils/cser"
)

func encodeSig(v, r, s *big.Int) (sig [65]byte) {
	copy(sig[0:], cser.PaddedBytes(r.Bytes(), 32)[:32])
	copy(sig[32:], cser.PaddedBytes(s.Bytes(), 32)[:32])
	copy(sig[64:], cser.PaddedBytes(v.Bytes(), 1)[:1])
	return sig
}

func decodeSig(sig [65]byte) (v, r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64]})
	return
}

func TransactionMarshalCSER(w *cser.Writer, tx *types.Transaction) error {
	w.U64(tx.Nonce())
	w.U64(tx.Gas())
	w.BigInt(tx.GasPrice())
	w.BigInt(tx.Value())
	w.Bool(tx.To() != nil)
	if tx.To() != nil {
		w.FixedBytes(tx.To().Bytes())
	}
	w.SliceBytes(tx.Data())
	sig := encodeSig(tx.RawSignatureValues())
	w.FixedBytes(sig[:])
	return nil
}

func TransactionUnmarshalCSER(r *cser.Reader) (*types.Transaction, error) {
	nonce := r.U64()
	gasLimit := r.U64()
	gasPrice := r.BigInt()
	amount := r.BigInt()
	toExists := r.Bool()
	var to *common.Address
	if toExists {
		var _to common.Address
		r.FixedBytes(_to[:])
		to = &_to
	}
	data := r.SliceBytes()
	// sig
	var sig [65]byte
	r.FixedBytes(sig[:])

	v, _r, s := decodeSig(sig)
	return types.NewRawTransaction(nonce, to, amount, gasLimit, gasPrice, data, v, _r, s), nil
}
