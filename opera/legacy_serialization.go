package opera

import (
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
)

type GasRulesRLPV0 struct {
	MaxEventGas  uint64
	EventGas     uint64
	ParentGas    uint64
	ExtraDataGas uint64
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

// EncodeRLP is for RLP serialization.
func (u Upgrades) EncodeRLP(w io.Writer) error {
	bitmap := struct {
		V uint64
	}{}
	if u.Berlin {
		bitmap.V |= berlinBit
	}
	if u.London {
		bitmap.V |= londonBit
	}
	if u.Llr {
		bitmap.V |= llrBit
	}
	return rlp.Encode(w, &bitmap)
}

// DecodeRLP is for RLP serialization.
func (u *Upgrades) DecodeRLP(s *rlp.Stream) error {
	bitmap := struct {
		V uint64
	}{}
	err := s.Decode(&bitmap)
	if err != nil {
		return err
	}
	u.Berlin = (bitmap.V & berlinBit) != 0
	u.London = (bitmap.V & londonBit) != 0
	u.Llr = (bitmap.V & llrBit) != 0
	return nil
}

// EncodeRLP is for RLP serialization.
func (r GasRules) EncodeRLP(w io.Writer) error {
	// write the type
	rType := uint8(0)
	if r.EpochVoteGas != 0 || r.MisbehaviourProofGas != 0 || r.BlockVotesBaseGas != 0 || r.BlockVoteGas != 0 {
		rType = 1
		_, err := w.Write([]byte{rType})
		if err != nil {
			return err
		}
	}
	if rType == 0 {
		return rlp.Encode(w, &GasRulesRLPV0{
			MaxEventGas:  r.MaxEventGas,
			EventGas:     r.EventGas,
			ParentGas:    r.ParentGas,
			ExtraDataGas: r.ExtraDataGas,
		})
	} else {
		return rlp.Encode(w, (*GasRulesRLPV1)(&r))
	}
}

// DecodeRLP is for RLP serialization.
func (r *GasRules) DecodeRLP(s *rlp.Stream) error {
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
	if rType == 0 {
		rlpR := GasRulesRLPV0{}
		err = s.Decode(&rlpR)
		if err != nil {
			return err
		}
		*r = GasRules{
			MaxEventGas:  rlpR.MaxEventGas,
			EventGas:     rlpR.EventGas,
			ParentGas:    rlpR.ParentGas,
			ExtraDataGas: rlpR.ExtraDataGas,
		}
		return nil
	} else {
		return s.Decode((*GasRulesRLPV1)(r))
	}
}
