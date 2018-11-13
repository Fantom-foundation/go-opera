// +build multi

// This is version of main.go with CPU profiling enabled.
// Use it in go build command instead of main.go
// lachesis.prof file is created in current directory with profiling data after execution finishes
// see https://blog.golang.org/profiling-go-programs for details
//
// TODO: add memory profiling when needed
// see https://golang.org/pkg/runtime/pprof/
//
package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"

	cmd "github.com/andrecronje/lachesis/cmd/lachesis/commands"
)

// TODO: Change so that this is a flag in default main and not auto startup

func main() {
	rootCmd := cmd.RootCmd

	rootCmd.AddCommand(
		cmd.VersionCmd,
		cmd.NewKeygenCmd(),
		cmd.NewRunCmd())

	//Do not print usage when error occurs
	rootCmd.SilenceUsage = true

	// Set up rpofiling
	f, err := os.Create("./lachesis.prof")
	if err != nil {
		os.Exit(127)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("captured %v, stopping profiler and exiting..\n", sig)
			pprof.StopCPUProfile()
			os.Exit(1)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
