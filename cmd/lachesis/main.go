package main

import (
	_ "net/http/pprof"
	"os"

	cmd "github.com/andrecronje/lachesis/cmd/lachesis/commands"
)

func main() {
	rootCmd := cmd.RootCmd

	rootCmd.AddCommand(
		cmd.VersionCmd,
		cmd.NewKeygenCmd(),
		cmd.NewRunCmd())

	//Do not print usage when error occurs
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
