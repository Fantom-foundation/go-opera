package launcher

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore"
)

func addFakeValidatorKey(ctx *cli.Context, key *ecdsa.PrivateKey, pubkey validatorpk.PubKey, valKeystore valkeystore.RawKeystoreI) {
	// add fake validator key
	if key != nil && !valKeystore.Has(pubkey) {
		err := valKeystore.Add(pubkey, crypto.FromECDSA(key), validatorpk.FakePassword)
		if err != nil {
			utils.Fatalf("Failed to add fake validator key: %v", err)
		}
	}
}

func getValKeystoreDir(cfg node.Config) string {
	keydir, err := cfg.KeyDirConfig()
	if err != nil {
		utils.Fatalf("Failed to setup account config: %v", err)
	}
	return keydir
}

// makeValidatorPasswordList reads password lines from the file specified by the global --validator.password flag.
func makeValidatorPasswordList(ctx *cli.Context) []string {
	if path := ctx.String(validatorPasswordFlag.Name); path != "" {
		text, err := ioutil.ReadFile(path)
		if err != nil {
			utils.Fatalf("Failed to read password file: %v", err)
		}
		lines := strings.Split(string(text), "\n")
		// Sanitise DOS line endings.
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], "\r")
		}
		return lines
	}
	if ctx.IsSet(FakeNetFlag.Name) {
		return []string{validatorpk.FakePassword}
	}
	return nil
}

func unlockValidatorKey(ctx *cli.Context, pubKey validatorpk.PubKey, valKeystore valkeystore.KeystoreI) error {
	if !valKeystore.Has(pubKey) {
		return valkeystore.ErrNotFound
	}
	var err error
	for trials := 0; trials < 3; trials++ {
		prompt := fmt.Sprintf("Unlocking validator key %s | Attempt %d/%d", pubKey.String(), trials+1, 3)
		password := getPassPhrase(prompt, false, 0, makeValidatorPasswordList(ctx))
		err = valKeystore.Unlock(pubKey, password)
		if err == nil {
			log.Info("Unlocked validator key", "pubkey", pubKey.String())
			return nil
		}
		if err.Error() != "could not decrypt key with given password" {
			return err
		}
	}
	// All trials expended to unlock account, bail out
	return err
}
