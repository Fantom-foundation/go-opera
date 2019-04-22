package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewID initialize id command.
func NewID() *cobra.Command {
	cmd := cobra.Command{
		Use:   "id",
		Short: "Prints node id",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("id")
			return nil
		},
	}

	return &cmd
}
