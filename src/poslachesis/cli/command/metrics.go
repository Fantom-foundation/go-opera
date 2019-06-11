package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Metrics turn on metrics.
var Metrics = &cobra.Command{
	Use: "metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return errors.New("not support arguments")
		}

		return nil
	},
}
