package command

import (
	"context"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
)

var Stake *cobra.Command

func init() {
	Stake = prepareStake()
}

// newStake prepares stake command.
func prepareStake() *cobra.Command {
	cmd := cobra.Command{
		Use:   "stake",
		Short: "Prints node stake",
	}

	var port int
	cmd.Flags().IntVarP(&port, "port", "p", managementPort, "lachesis management port")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(port)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
		defer cancel()

		req := empty.Empty{}
		resp, err := client.Stake(ctx, &req)
		if err != nil {
			return err
		}

		cmd.Println(resp.Value)
		return nil
	}

	return &cmd
}
