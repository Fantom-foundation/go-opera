package command

import (
	"github.com/spf13/cobra"
)

// ID prints id of the node.
var ID = &cobra.Command{
	Use:   "id",
	Short: "Prints id of the node",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		id, err := proxy.GetSelfID()
		if err != nil {
			return err
		}

		cmd.Println(id.Hex())
		return nil
	},
}

func init() {
	initCtrlProxy(ID)
}
