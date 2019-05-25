package command

import (
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Balance prints stake balance of a peer.
var Balance = &cobra.Command{
	Use:   "balance",
	Short: "Prints stake balance of a peer",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		var id hash.Peer
		hex, err := cmd.Flags().GetString("peer")
		if err != nil || hex == "self" {
			id, err = proxy.GetSelfID()
		} else {
			id = hash.HexToPeer(hex)
		}
		if err != nil {
			return err
		}

		balance, err := proxy.GetBalanceOf(id)
		if err != nil {
			return err
		}

		cmd.Printf("balance of %s == %d\n", id.Hex(), balance)
		return nil
	},
}

func init() {
	initCtrlProxy(Balance)

	Balance.Flags().String("peer", "self", "peer ID")
}
