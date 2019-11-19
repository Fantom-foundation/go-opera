package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/params"
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
	flags = []cli.Flag{
		NumberFlag,
		PeriodFlag,
		TxnsRateFlag,
		utils.MetricsEnabledFlag,
		MetricsPrometheusEndpointFlag,
		VerbosityFlag,
	}

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
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	args := ctx.Args()
	if len(args) != 1 {
		return fmt.Errorf("url expected")
	}

	SetupPrometheus(ctx)

	url := args[0]
	num, ofTotal := getNumber(ctx)
	maxTxnsPerSec := getTxnsRate(ctx)
	period := getPeriod(ctx)

	tt := newThreads(url, num, ofTotal, maxTxnsPerSec, period)
	tt.SetName("Threads")
	tt.Start()

	waitForSignal()
	tt.Stop()
	return nil
}

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
}
