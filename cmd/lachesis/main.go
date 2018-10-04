package main

import (
	_ "net/http/pprof"

	cmd "github.com/andrecronje/lachesis/cmd/lachesis/commands"
)

func main() {
	rootCmd := cmd.RootCmd

	rootCmd.AddCommand(
		cmd.VersionCmd,
		cmd.KeygenCmd,
		cmd.NewRunCmd())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
