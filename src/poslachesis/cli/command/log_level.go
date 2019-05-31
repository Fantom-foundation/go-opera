package command

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	// ErrOneArgument returns when command expects exactly one
	// argument.
	ErrOneArgument = errors.New("expected exactly one argument")
)

// LogLevel sets logger log level.
var LogLevel = &cobra.Command{
	Use:   "log-level",
	Short: "Sets logger log level",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		if len(args) != 1 {
			return ErrOneArgument
		}

		if err := proxy.SetLogLevel(args[0]); err != nil {
			return err
		}

		cmd.Println("ok")
		return nil
	},
}

func init() {
	initCtrlProxy(LogLevel)
}
