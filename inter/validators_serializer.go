package inter

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Fantom-foundation/go-opera/inter/drivertype"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
)

// RPCMarshalValidators converts the given validators to the RPC output .
func RPCMarshalValidators(validators ValidatorProfiles) map[hexutil.Uint64]interface{} {
	res := make(map[hexutil.Uint64]interface{}, len(validators))
	for vid, profile := range validators {
		res[hexutil.Uint64(vid)] = map[string]interface{}{
			"weight": (*hexutil.Big)(profile.Weight),
			"pubkey": profile.PubKey.String(),
		}
	}

	return res
}

// RPCUnmarshalValidators converts the RPC output to the validators.
func RPCUnmarshalValidators(fields map[hexutil.Uint64]interface{}) (ValidatorProfiles, error) {
	validators := make(ValidatorProfiles, len(fields))

	for vid, val := range fields {
		profile := val.(map[string]interface{})
		pk, err := validatorpk.FromString(profile["pubkey"].(string))
		if err != nil {
			return nil, err
		}
		validators[idx.ValidatorID(vid)] = drivertype.Validator{
			Weight: profile["weight"].(*big.Int),
			PubKey: pk,
		}
	}

	return validators, nil
}
