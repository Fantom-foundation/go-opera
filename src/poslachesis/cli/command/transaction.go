package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Transaction makes a transaction for stake transfer.
var Transaction = &cobra.Command{
	Use:   "transaction",
	Short: "Transaction returns information amount tranation",
	RunE: func(cmd *cobra.Command, args []string) error {
		hex, err := cmd.Flags().GetString("hex")
		if err != nil {
			return err
		}

		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		tx, err := proxy.GetTransaction(hash.HexToTransactionHash(hex))
		if err != nil {
			return err
		}

		message := fmt.Sprintf(
			"transfer %d from %s to %s %s",
			tx.Amount,
			tx.Sender.Hex(),
			tx.Receiver.Hex(),
			confirmedToHuman(tx.Confirmed),
		)

		cmd.Println(message)
		return nil
	},
}

func confirmedToHuman(c bool) string {
	if c {
		return "cofirmed"
	}

	return "unconfirmed"
}

func init() {
	initCtrlProxy(Transaction)

	Transaction.Flags().String("hex", "", "transaction hex (required)")

	if err := Transaction.MarkFlagRequired("hex"); err != nil {
		panic(err)
	}
}
