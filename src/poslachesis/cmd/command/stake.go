package command

import (
	"context"
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewStake initialize stake command.
func NewStake() *cobra.Command {
	cmd := cobra.Command{
		Use:   "stake",
		Short: "Prints node stake",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := proxy.NewManagementClient("localhost:55557")
			if err != nil {
				return errors.Wrap(err, "create client")
			}

			ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
			defer cancel()

			req := empty.Empty{}
			resp, err := cli.Stake(ctx, &req)
			if err != nil {
				return errors.Wrap(err, "get id")
			}

			fmt.Println(resp.Value)
			return nil
		},
	}

	return &cmd
}
