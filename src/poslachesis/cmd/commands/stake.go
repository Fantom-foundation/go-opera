package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewStake initialize stake command.
func NewStake() *cobra.Command {
	cmd := cobra.Command{
		Use:   "stake",
		Short: "Prints node stake",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("stake")
			return nil
		},
	}

	return &cmd
}
