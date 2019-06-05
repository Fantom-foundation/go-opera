package command

import (
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// TxnInfo returns information about transaction.
var TxnInfo = &cobra.Command{
	Use:   "txn",
	Short: "returns information about transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		for _, hex := range args {
			tx, event, block, err := proxy.GetTxnInfo(hash.HexToTransactionHash(hex))
			if err != nil {
				return err
			}

			defer cmd.Printf("\n")

			if tx == nil {
				cmd.Printf("transfer %s not found", hex)
				return nil
			}
			cmd.Printf("transfer %d to %s",
				tx.Amount,
				tx.Receiver.Hex(),
			)

			if event == nil {
				cmd.Printf(" is in mempool")
				return nil
			}
			cmd.Printf(" included into event %s",
				event.Hash().Hex(),
			)

			if block == nil {
				cmd.Printf(" and is not confirmed yet")
				return nil
			}
			cmd.Printf(" and is confirmed by block %d",
				block.Index,
			)
		}

		return nil
	},
}

func init() {
	initCtrlProxy(TxnInfo)
}
