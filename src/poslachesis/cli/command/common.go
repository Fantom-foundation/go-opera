package command

import (
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/spf13/cobra"
)

func initCtrlProxy(cmd *cobra.Command) {
	cmd.Flags().String("addr", "localhost:55557", "node control net addr")
}

func makeCtrlProxy(cmd *cobra.Command) (proxy.NodeProxy, error) {
	addr, err := cmd.Flags().GetString("addr")
	if err != nil {
		return nil, err
	}

	grpcProxy, err := proxy.NewGrpcNodeProxy(addr, nil)
	if err != nil {
		return nil, err
	}

	return grpcProxy, nil
}
