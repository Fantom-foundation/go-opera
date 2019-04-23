package command

import (
	"context"
	"fmt"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

const (
	connTimeout = 3 * time.Second
)

// NewID initialize id command.
func NewID() *cobra.Command {
	cmd := cobra.Command{
		Use:   "id",
		Short: "Prints node id",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := proxy.NewManagementClient("localhost:55557")
			if err != nil {
				return errors.Wrap(err, "create client")
			}

			ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
			defer cancel()

			req := empty.Empty{}
			resp, err := cli.ID(ctx, &req)
			if err != nil {
				return errors.Wrap(err, "get id")
			}

			fmt.Println(resp.Id)
			return nil
		},
	}

	return &cmd
}
