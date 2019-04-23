package command

import (
	"context"
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
	"github.com/spf13/cobra"
)

const (
	internalTxAddedMessage = "Internal transaction has been added"
)

// NewTx initialie internal transaction create command.
func NewTx() *cobra.Command {
	cmd := cobra.Command{
		Use:   "tx",
		Short: "Adds internal transaction",
	}

	var amount uint64
	var port int
	var to string

	cmd.Flags().Uint64VarP(&amount, "amount", "", 0, "internal transaction amount (required)")
	cmd.Flags().IntVarP(&port, "port", "p", 55557, "lachesis management port")
	cmd.Flags().StringVarP(&to, "to", "", "", "destination node (required)")
	cmd.MarkFlagRequired("amount")
	cmd.MarkFlagRequired("to")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("localhost:%d", port)
		cli, err := proxy.NewManagementClient(addr)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
		defer cancel()

		req := proto.InternalTxRequest{
			Amount:   amount,
			Receiver: to,
		}
		if _, err := cli.InternalTx(ctx, &req); err != nil {
			return err
		}

		fmt.Println(internalTxAddedMessage)
		return nil
	}

	return &cmd

}
