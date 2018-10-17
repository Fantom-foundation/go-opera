package main

import (
	_ "net/http/pprof"
	"os"
 	cmd "github.com/andrecronje/lachesis/cmd/dummy/commands"
)

func main() {
	rootCmd := cmd.RootCmd
 	//Do not print usage when error occurs
	rootCmd.SilenceUsage = true
 	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
