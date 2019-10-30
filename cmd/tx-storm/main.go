package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/params"
)

const (
	// clientIdentifier to advertise over the network.
	clientIdentifier = "txn-storm"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags).
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, gitDate, "the transactions generator CLI")

	flags []cli.Flag
)

// init the CLI app.
func init() {

	// Flags.
	flags = []cli.Flag{}

	// App.
	app.Action = generatorMain
	app.Version = params.VersionWithCommit(gitCommit, gitDate)

	app.Commands = []cli.Command{}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, flags...)

	app.Before = func(ctx *cli.Context) error {
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generatorMain is the main entry point.
func generatorMain(ctx *cli.Context) error {
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	return nil
}
