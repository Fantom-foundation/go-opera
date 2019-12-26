package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/params"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	_ "github.com/Fantom-foundation/go-lachesis/version"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags).
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, gitDate, "the fake account generator CLI")

	flags = []cli.Flag{
		FromFlag,
		CountFlag,
	}
)

// init the CLI app.
func init() {
	app.Action = generatorMain
	app.Version = params.VersionWithCommit(gitCommit, gitDate)

	app.Commands = []cli.Command{}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, flags...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generatorMain is the main entry point.
func generatorMain(ctx *cli.Context) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	from := getFrom(ctx)
	count := getCount(ctx)

	balance := genesis.Account{
		Balance: big.NewInt(1e18),
	}
	jsonData, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	comma := ""
	fmt.Fprint(os.Stdout, "{")
	defer fmt.Fprintf(os.Stdout, `}`)
	for addr := range NewAccs(from, count, sigs) {
		fmt.Fprintf(os.Stdout, "%s%q: %s\n", comma, addr.Hex(), string(jsonData))
		comma = ","
	}

	return nil
}
