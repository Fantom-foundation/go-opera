package main

import (
	"fmt"
	"os"
	"reflect"
	"unicode"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       nodeFlags,
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func makeLachesisConfig(ctx *cli.Context) lachesis.Config {
	cfg, _, _ := lachesis.FakeNetConfig(0)
	// TODO: apply flags
	return cfg
}

func makeGossipConfig(ctx *cli.Context, network lachesis.Config) gossip.Config {
	cfg := gossip.DefaultConfig(network)
	// TODO: apply flags

	return cfg
}

func makeNodeConfig(ctx *cli.Context) *node.Config {
	cfg := defaultNodeConfig()

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg)

	return &cfg
}

func defaultNodeConfig() node.Config {
	cfg := NodeDefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "lachesis")
	cfg.WSModules = append(cfg.WSModules, "lachesis")
	cfg.IPCPath = "lachesis.ipc"
	cfg.P2P.DiscoveryV5 = true
	return cfg
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	cfg := makeNodeConfig(ctx)
	comment := ""

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}
