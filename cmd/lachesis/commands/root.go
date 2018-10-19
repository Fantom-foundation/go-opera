package commands
 import (
	"github.com/spf13/cobra"
)

//RootCmd is the root command for Lachesis
var RootCmd = &cobra.Command{
	Use:              "lachesis",
	Short:            "lachesis consensus",
	TraverseChildren: true,
}
