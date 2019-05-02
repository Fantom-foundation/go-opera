package command

import (
	"context"

	"github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
	"github.com/spf13/cobra"
)

const (
	internalTxnAddedMsg = "Internal transaction has been added"
)

var InternalTxn = &cobra.Command{
	Use:   "internal_txn",
	Short: "Adds internal transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
		defer cancel()

		amount, err := cmd.Flags().GetUint64("amount")
		if err != nil {
			return err
		}
		receiver, err := cmd.Flags().GetString("receiver")
		if err != nil {
			return err
		}
		req := wire.InternalTxnRequest{
			Amount:   amount,
			Receiver: receiver,
		}
		if _, err := client.InternalTxn(ctx, &req); err != nil {
			return err
		}

		cmd.Println(internalTxnAddedMsg)
		return nil
	},
}

func init() {
	InternalTxn.Flags().String("receiver", "", "transaction receiver (required)")
	InternalTxn.Flags().Uint64("amount", 0, "transaction amount (required)")
	InternalTxn.MarkFlagRequired("receiver")
	InternalTxn.MarkFlagRequired("amount")
}
