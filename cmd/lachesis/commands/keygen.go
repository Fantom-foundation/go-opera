package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

var (
	privKeyFile           string
	pubKeyFile            string
	config                = NewDefaultCLIConfig()
	defaultPrivateKeyFile = fmt.Sprintf("%s/priv_key.pem", config.Lachesis.DataDir)
	defaultPublicKeyFile  = fmt.Sprintf("%s/key.pub", config.Lachesis.DataDir)
)

// NewKeygenCmd produces a KeygenCmd which creates a key pair
func NewKeygenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Create new key pair",
		RunE:  keygen,
	}
	AddKeygenFlags(cmd)
	return cmd
}

//AddKeygenFlags adds flags to the keygen command
func AddKeygenFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&privKeyFile, "pem", defaultPrivateKeyFile, "File where the private key will be written")
	cmd.Flags().StringVar(&pubKeyFile, "pub", defaultPublicKeyFile, "File where the public key will be written")
}
func keygen(cmd *cobra.Command, args []string) error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("error generating private key")
	}
	if err := os.MkdirAll(path.Dir(privKeyFile), 0700); err != nil {
		return fmt.Errorf("writing private key: %v", err)
	}
	_, err = os.Stat(privKeyFile)
	if err == nil {
		return fmt.Errorf("A key already lives under: %s", path.Dir(privKeyFile))
	}
	privFile, err := os.Create(privKeyFile)
	if err != nil {
		return fmt.Errorf("open private key: %v", err)
	}
	if err := key.WriteTo(privFile); err != nil {
		return fmt.Errorf("writing private key: %v", err)
	}
	fmt.Printf("Your private key has been saved to: %s\n", privKeyFile)

	if err := os.MkdirAll(path.Dir(pubKeyFile), 0700); err != nil {
		return fmt.Errorf("writing public key: %v", err)
	}
	pub := fmt.Sprintf("0x%X", key.Public().Bytes())
	if err := ioutil.WriteFile(pubKeyFile, []byte(pub), 0666); err != nil {
		return fmt.Errorf("writing public key: %v", err)
	}
	fmt.Printf("Your public key has been saved to: %s\n", pubKeyFile)
	return nil
}
