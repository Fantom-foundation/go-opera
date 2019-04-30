package command

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
)

var ID = &cobra.Command{
	Use:   "id",
	Short: "Prints node id",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
		defer cancel()

		req := empty.Empty{}
		resp, err := client.ID(ctx, &req)
		if err != nil {
			return err
		}

		cmd.Println(resp.Id)
		return nil
	},
}
