package inter

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/utils/cser"
)

var ErrUnknownTxType = errors.New("unknown tx type")

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
	if tx.Type() != types.LegacyTxType && tx.Type() != types.AccessListTxType && tx.Type() != types.DynamicFeeTxType && tx.Type() != types.AccountAbstractionTxType {
		return ErrUnknownTxType
	}
	if tx.Type() != types.LegacyTxType {
		// marker of a non-standard tx
		w.BitsW.Write(6, 0)
		// tx type
		w.U8(tx.Type())
	} else if tx.Gas() <= 0xff {
		return errors.New("cannot serialize legacy tx with gasLimit <= 256")
	}
	w.U64(tx.Nonce())
	w.U64(tx.Gas())
	if tx.Type() == types.DynamicFeeTxType {
		w.BigInt(tx.GasTipCap())
		w.BigInt(tx.GasFeeCap())
	} else {
		w.BigInt(tx.GasPrice())
	}
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
	if tx.Type() == types.AccessListTxType || tx.Type() == types.DynamicFeeTxType {
		w.BigInt(tx.ChainId())
		w.U32(uint32(len(tx.AccessList())))
		for _, tuple := range tx.AccessList() {
			w.FixedBytes(tuple.Address.Bytes())
			w.U32(uint32(len(tuple.StorageKeys)))
			for _, h := range tuple.StorageKeys {
				w.FixedBytes(h.Bytes())
			}
		}
	}
	return nil
}

func TransactionUnmarshalCSER(r *cser.Reader) (*types.Transaction, error) {
	txType := uint8(types.LegacyTxType)
	if r.BitsR.View(6) == 0 {
		r.BitsR.Read(6)
		txType = r.U8()
	}

	nonce := r.U64()
	gasLimit := r.U64()
	var gasPrice *big.Int
	var gasTipCap *big.Int
	var gasFeeCap *big.Int
	if txType == types.DynamicFeeTxType {
		gasTipCap = r.BigInt()
		gasFeeCap = r.BigInt()
	} else {
		gasPrice = r.BigInt()
	}
	amount := r.BigInt()
	toExists := r.Bool()
	var to *common.Address
	if toExists {
		var _to common.Address
		r.FixedBytes(_to[:])
		to = &_to
	}
	data := r.SliceBytes(ProtocolMaxMsgSize)
	// sig
	v := r.BigInt()
	var sig [64]byte
	r.FixedBytes(sig[:])
	_r, s := decodeSig(sig)

	if txType == types.LegacyTxType {
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
	} else if txType == types.AccessListTxType || txType == types.DynamicFeeTxType {
		chainID := r.BigInt()
		accessListLen := r.U32()
		if accessListLen > ProtocolMaxMsgSize/24 {
			return nil, cser.ErrTooLargeAlloc
		}
		accessList := make(types.AccessList, accessListLen)
		for i := range accessList {
			r.FixedBytes(accessList[i].Address[:])
			keysLen := r.U32()
			if keysLen > ProtocolMaxMsgSize/32 {
				return nil, cser.ErrTooLargeAlloc
			}
			accessList[i].StorageKeys = make([]common.Hash, keysLen)
			for j := range accessList[i].StorageKeys {
				r.FixedBytes(accessList[i].StorageKeys[j][:])
			}
		}
		if txType == types.AccessListTxType {
			return types.NewTx(&types.AccessListTx{
				ChainID:    chainID,
				Nonce:      nonce,
				GasPrice:   gasPrice,
				Gas:        gasLimit,
				To:         to,
				Value:      amount,
				Data:       data,
				AccessList: accessList,
				V:          v,
				R:          _r,
				S:          s,
			}), nil
		} else {
			return types.NewTx(&types.DynamicFeeTx{
				ChainID:    chainID,
				Nonce:      nonce,
				GasTipCap:  gasTipCap,
				GasFeeCap:  gasFeeCap,
				Gas:        gasLimit,
				To:         to,
				Value:      amount,
				Data:       data,
				AccessList: accessList,
				V:          v,
				R:          _r,
				S:          s,
			}), nil
		}
	}
	return nil, ErrUnknownTxType
}
