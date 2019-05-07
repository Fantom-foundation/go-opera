package command

import (
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/spf13/cobra"
)

// ID prints id of the node.
var ID = &cobra.Command{
	Use:   "id",
	Short: "Prints node id",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := proxy.NewGrpcCmdProxy(ctrlAddr, connTimeout)
		if err != nil {
			return err
		}

		id, err := proxy.GetID()
		if err != nil {
			return err
		}

		cmd.Println(id)
		return nil
	},
}
