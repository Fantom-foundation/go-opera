package encryption

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
)

type encryptedAccountKeyJSONV3 struct {
	Address string              `json:"address"`
	Crypto  keystore.CryptoJSON `json:"crypto"`
	Id      string              `json:"id"`
	Version int                 `json:"version"`
}

func MigrateAccountToValidatorKey(acckeypath string, valkeypath string, pubkey validatorpk.PubKey) error {
	acckeyjson, err := ioutil.ReadFile(acckeypath)
	if err != nil {
		return err
	}
	acck := new(encryptedAccountKeyJSONV3)
	if err := json.Unmarshal(acckeyjson, acck); err != nil {
		return err
	}

	valk := EncryptedKeyJSON{
		Type:      validatorpk.Types.Secp256k1,
		PublicKey: common.Bytes2Hex(pubkey.Raw),
		Crypto:    acck.Crypto,
	}
	valkeyjson, err := json.Marshal(valk)
	if err != nil {
		return err
	}
	tmpName, err := writeTemporaryKeyFile(valkeypath, valkeyjson)
	if err != nil {
		return err
	}
	return os.Rename(tmpName, valkeypath)
}
