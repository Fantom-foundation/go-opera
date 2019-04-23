package command

import (
	"github.com/spf13/cobra"
)

// NewInternal initialize internal commands.
func NewInternal() *cobra.Command {
	cmd := cobra.Command{
		Use:   "internal",
		Short: "Internal commands",
	}

	txCmd := NewTx()
	cmd.AddCommand(txCmd)

	return &cmd
}
