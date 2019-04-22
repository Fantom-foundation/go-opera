package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/poslachesis/cmd/commands"
)

func main() {
	rootCmd := &cobra.Command{
		Use: os.Args[0],
	}

	rootCmd.AddCommand(commands.NewInternal())
	rootCmd.AddCommand(commands.NewID())
	rootCmd.AddCommand(commands.NewStake())
	rootCmd.AddCommand(commands.NewStart())

	rootCmd.Execute()
}
