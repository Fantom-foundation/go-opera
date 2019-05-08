package command

import (
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Transfer makes transaction to stake transfer.
var Transfer = &cobra.Command{
	Use:   "transfer",
	Short: "Adds internal transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := cmd.Flags().GetUint64("amount")
		if err != nil {
			return err
		}
		hex, err := cmd.Flags().GetString("receiver")
		if err != nil {
			return err
		}
		receiver := hash.HexToPeer(hex)

		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		err = proxy.SendTo(amount, receiver)
		if err != nil {
			return err
		}

		cmd.Println("ok")
		return nil
	},
}

func init() {
	initCtrlProxy(Transfer)

	Transfer.Flags().String("receiver", "", "transaction receiver (required)")
	Transfer.Flags().Uint64("amount", 0, "transaction amount (required)")

	if err := Transfer.MarkFlagRequired("receiver"); err != nil {
		panic(err)
	}
	if err := Transfer.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
}
