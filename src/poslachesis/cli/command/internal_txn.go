package command

import (
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/spf13/cobra"
)

const (
	internalTxnAddedMsg = "internal transaction has been added"
)

// InternalTxn adds internal transaction into the node.
var InternalTxn = &cobra.Command{
	Use:   "internal_txn",
	Short: "Adds internal transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := proxy.NewGrpcCmdProxy(ctrlAddr, connTimeout)
		if err != nil {
			return err
		}

		amount, err := cmd.Flags().GetUint64("amount")
		if err != nil {
			return err
		}
		receiver, err := cmd.Flags().GetString("receiver")
		if err != nil {
			return err
		}

		if err := proxy.SubmitInternalTxn(amount, receiver); err != nil {
			return err
		}

		cmd.Println(internalTxnAddedMsg)
		return nil
	},
}

func init() {
	// TODO: move to command scope
	InternalTxn.Flags().String("receiver", "", "transaction receiver (required)")
	InternalTxn.Flags().Uint64("amount", 0, "transaction amount (required)")

	if err := InternalTxn.MarkFlagRequired("receiver"); err != nil {
		panic(err)
	}
	if err := InternalTxn.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
}
