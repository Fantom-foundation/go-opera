package command

import (
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Info returns information about transaction.
var Info = &cobra.Command{
	Use:   "info",
	Short: "Info returns information about transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		for _, hex := range args {
			tx, err := proxy.GetTransaction(hash.HexToTransactionHash(hex))
			if err != nil {
				return err
			}

			if tx == nil {
				cmd.Printf("transfer %s not found\n", hex)
				return nil
			}

			cmd.Printf("transfer %d to %s\n",
				tx.Amount,
				tx.Receiver.Hex(),
			)
		}

		return nil
	},
}

func init() {
	initCtrlProxy(Info)
}
