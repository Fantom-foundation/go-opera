package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/crypto"
)

func TestFakeNetFlag_NonValidator(t *testing.T) {
	// Start a lachesis console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--fakenet", "0/3",
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"console")

	// Gather all the infos the welcome message needs to contain
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

	waitForEndpoint(t, filepath.Join(cli.Datadir, "lachesis.ipc"), 10*time.Second)

	// Verify the actual welcome message to the required template
	cli.Expect(`
Welcome to the Lachesis JavaScript console!

instance: go-lachesis/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Coinbase}}
at block: 0 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

> {{.InputLine "exit"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked fake validator",
	}
	for _, m := range wantMessages {
		if strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text contains %q", m)
		}
	}
}

func TestFakeNetFlag_Validator(t *testing.T) {
	// Start a lachesis console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--fakenet", "3/3",
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"console")

	// Gather all the infos the welcome message needs to contain
	va := readFakeValidator("3/3")
	cli.Coinbase = strings.ToLower(va.Hex())
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

	waitForEndpoint(t, filepath.Join(cli.Datadir, "lachesis.ipc"), 10*time.Second)

	// Verify the actual welcome message to the required template
	cli.Expect(`
Welcome to the Lachesis JavaScript console!

instance: go-lachesis/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Coinbase}}
at block: 0 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

> {{.InputLine "exit"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked fake validator",
		"=" + va.Hex(),
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func readFakeValidator(fakenet string) *common.Address {
	n, _, err := parseFakeGen(fakenet)
	if err != nil {
		panic(err)
	}

	if n < 1 {
		return nil
	}

	addr := crypto.PubkeyToAddress(crypto.FakeKey(n).PublicKey)
	return &addr
}
