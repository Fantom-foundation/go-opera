package launcher

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"path"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/valkeystore"
)

var (
	validatorCommand = cli.Command{
		Name:     "validator",
		Usage:    "Manage validators",
		Category: "VALIDATOR COMMANDS",
		Description: `

Create a new validator private key.

It supports interactive mode, when you are prompted for password as well as
non-interactive mode where passwords are supplied via a given password file.
Non-interactive mode is only meant for scripted use on test networks or known
safe environments.

Make sure you remember the password you gave when creating a new validator key.
Without it you are not able to unlock your validator key.

Note that exporting your key in unencrypted format is NOT supported.

Keys are stored under <DATADIR>/keystore/validator.
It is safe to transfer the entire directory or the individual keys therein
between Opera nodes by simply copying.

Make sure you backup your keys regularly.`,
		Subcommands: []cli.Command{
			{
				Name:   "new",
				Usage:  "Create a new validator key",
				Action: utils.MigrateFlags(validatorKeyCreate),
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.KeyStoreDirFlag,
					utils.PasswordFileFlag,
				},
				Description: `
    opera validator new

Creates a new validator private key and prints the public key.

The key is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your key in the future.

For non-interactive use the passphrase can be specified with the --validator.password flag:

Note, this is meant to be used for testing only, it is a bad idea to save your
password to file or expose in any other way.
`,
			},
		},
	}
)

// validatorKeyCreate creates a new validator key into the keystore defined by the CLI flags.
func validatorKeyCreate(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)
	utils.SetNodeConfig(ctx, &cfg.Node)

	password := getPassPhrase("Your new validator key is locked with a password. Please give a password. Do not forget this password.", true, 0, utils.MakePasswordList(ctx))

	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
	}
	privateKey := crypto.FromECDSA(privateKeyECDSA)
	publicKey := validator.PubKey{
		Raw:  crypto.FromECDSAPub(&privateKeyECDSA.PublicKey),
		Type: "secp256k1",
	}

	valKeystore := valkeystore.NewDefaultFileRawKeystore(path.Join(getValKeystoreDir(cfg.Node), "validator"))
	err = valKeystore.Add(publicKey, privateKey, password)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
	}

	// Sanity check
	_, err = valKeystore.Get(publicKey, password)
	if err != nil {
		utils.Fatalf("Failed to decrypt the account: %v", err)
	}

	fmt.Printf("\nYour new key was generated\n\n")
	fmt.Printf("Public key:                  %s\n", publicKey.String())
	fmt.Printf("Path of the secret key file: %s\n\n", valKeystore.PathOf(publicKey))
	fmt.Printf("- You can share your public key with anyone. Others need it to validate messages from you.\n")
	fmt.Printf("- You must NEVER share the secret key with anyone! The key controls access to your validator!\n")
	fmt.Printf("- You must BACKUP your key file! Without the key, it's impossible to operate the validator!\n")
	fmt.Printf("- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!\n\n")
	return nil
}
