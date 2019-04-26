package command

import (
	"context"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
)

var ID *cobra.Command

func init() {
	ID = prepareID()
}

// prepareID prepares id command.
func prepareID() *cobra.Command {
	cmd := cobra.Command{
		Use:   "id",
		Short: "Prints node id",
	}

	var port int
	cmd.Flags().IntVarP(&port, "port", "p", 55557, "lachesis management port")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(port)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
		defer cancel()

		req := empty.Empty{}
		resp, err := client.ID(ctx, &req)
		if err != nil {
			return err
		}

		cmd.Println(resp.Id)
		return nil
	}

	return &cmd
}
