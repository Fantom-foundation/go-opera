package command

import (
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Transfer makes a transaction for stake transfer.
var Transfer = &cobra.Command{
	Use:   "transfer",
	Short: "Transfers a balance amount to given receiver",
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := cmd.Flags().GetUint64("index")
		if err != nil {
			return err
		}
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

		h, err := proxy.SendTo(receiver, index, amount, 0)
		if err != nil {
			return err
		}

		cmd.Println(h.Hex())
		return nil
	},
}

func init() {
	initCtrlProxy(Transfer)

	Transfer.Flags().Uint64("index", 0, "transaction nonce (required)")
	Transfer.Flags().String("receiver", "", "transaction receiver (required)")
	Transfer.Flags().Uint64("amount", 0, "transaction amount (required)")

	if err := Transfer.MarkFlagRequired("index"); err != nil {
		panic(err)
	}
	if err := Transfer.MarkFlagRequired("receiver"); err != nil {
		panic(err)
	}
	if err := Transfer.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
}
