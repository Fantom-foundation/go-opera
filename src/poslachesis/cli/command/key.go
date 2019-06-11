package command

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

// Key generates private pem key.
var Key = &cobra.Command{
	Use:   "key",
	Short: "Generates private pem key",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		key, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		if err := key.WriteTo(file); err != nil {
			return err
		}

		cmd.Printf("%s created\n", filePath)
		return nil
	},
}

func init() {
	Key.Flags().String("file", "priv_key.pem", "file path to write")
}
