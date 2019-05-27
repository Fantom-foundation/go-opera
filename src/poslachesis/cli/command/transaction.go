package command

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

var (
	// ErrTooMuchArguments returns when cmd receives
	// too much arguments.
	ErrTooMuchArguments = errors.New("too much arguments")
	// ErrNotEnoughArguments returns when cmd receives
	// not enough arguments.
	ErrNotEnoughArguments = errors.New("not enough arguments")
)

// Transaction makes a transaction for stake transfer.
var Transaction = &cobra.Command{
	Use:   "transaction",
	Short: "Transaction returns information about transaction passed as first argument",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return ErrTooMuchArguments
		}
		if len(args) < 1 {
			return ErrNotEnoughArguments
		}

		proxy, err := makeCtrlProxy(cmd)
		if err != nil {
			return err
		}
		defer proxy.Close()

		tx, err := proxy.GetTransaction(hash.HexToTransactionHash(args[0]))
		if err != nil {
			return err
		}

		message := fmt.Sprintf(
			"transfer %d from %s to %s %s",
			tx.Amount,
			tx.Sender.Hex(),
			tx.Receiver.Hex(),
			confirmedToHuman(tx.Confirmed),
		)

		cmd.Println(message)
		return nil
	},
}

func confirmedToHuman(c bool) string {
	if c {
		return "cofirmed"
	}

	return "unconfirmed"
}

func init() {
	initCtrlProxy(Transaction)
}
