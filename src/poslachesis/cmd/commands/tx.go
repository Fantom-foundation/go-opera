package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewTx initialie internal transaction create command.
func NewTx() *cobra.Command {
	cmd := cobra.Command{
		Use:   "tx",
		Short: "Adds internal transaction",
	}

	var amount float64
	var to string

	cmd.Flags().Float64VarP(&amount, "amount", "", 0, "internal transaction amount (required)")
	cmd.Flags().StringVarP(&to, "to", "", "", "destination node (required)")
	cmd.MarkFlagRequired("amount")
	cmd.MarkFlagRequired("to")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println(amount, to)

		return nil
	}

	return &cmd

}
