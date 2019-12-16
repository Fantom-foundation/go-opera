package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/crypto"
)

func TestFakeNetFlag(t *testing.T) {
	// Start a lachesis console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--fakenet", "1/3",
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"console")

	// Gather all the infos the welcome message needs to contain
	cliSetFakeCoinbase(cli, "1/3")
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

	waitForEndpoint(t, filepath.Join(cli.Datadir, "lachesis.ipc"), 10*time.Second)

	// Verify the actual welcome message to the required template
	// TODO: clone (or PR) "github.com/ethereum/go-ethereum/console" to customize welcome message
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
		"Unlocked fake coinbase",
		"=0xF88D5892faF084DCF4143566d9C9b3F047c153Ca",
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func cliSetFakeCoinbase(cli *testcli, fakenet string) {
	n, _, err := parseFakeGen(fakenet)
	if err != nil {
		panic(err)
	}

	cli.Coinbase = strings.ToLower(
		crypto.PubkeyToAddress(crypto.FakeKey(n).PublicKey).Hex())
}
