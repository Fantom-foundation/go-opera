package commands

import (
	"github.com/spf13/cobra"
)

var (
	config = NewDefaultCLIConfig()
)

//RootCmd is the root command for Lachesis
var RootCmd = &cobra.Command{
	Use:              "lachesis",
	Short:            "lachesis consensus",
	TraverseChildren: true,
}
