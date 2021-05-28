package inter

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/utils/cser"
)

func encodeSig(r, s *big.Int) (sig [64]byte) {
	copy(sig[0:], cser.PaddedBytes(r.Bytes(), 32)[:32])
	copy(sig[32:], cser.PaddedBytes(s.Bytes(), 32)[:32])
	return sig
}

func decodeSig(sig [64]byte) (r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
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
	v, r, s := tx.RawSignatureValues()
	w.BigInt(v)
	sig := encodeSig(r, s)
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
	v := r.BigInt()
	var sig [64]byte
	r.FixedBytes(sig[:])
	_r, s := decodeSig(sig)

	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       to,
		Value:    amount,
		Data:     data,
		V:        v,
		R:        _r,
		S:        s,
	}), nil
}
