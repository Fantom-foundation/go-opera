package command

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
)

var Stake = &cobra.Command{
	Use:   "stake",
	Short: "Prints node stake",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
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
	},
}
