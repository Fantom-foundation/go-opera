package command

import (
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/spf13/cobra"
)

// Stake prints stake of the node.
var Stake = &cobra.Command{
	Use:   "stake",
	Short: "Prints node stake",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := proxy.NewGrpcCmdProxy(ctrlAddr, connTimeout)
		if err != nil {
			return err
		}

		stake, err := proxy.GetStake()
		if err != nil {
			return err
		}

		cmd.Println(stake)
		return nil
	},
}
