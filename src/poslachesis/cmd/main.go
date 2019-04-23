package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/poslachesis/cmd/command"
)

func main() {
	app := &cobra.Command{
		Use: os.Args[0],
	}

	app.AddCommand(command.NewInternal())
	app.AddCommand(command.NewID())
	app.AddCommand(command.NewStake())
	app.AddCommand(command.NewStart())

	app.Execute()
}
