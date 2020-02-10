package main

import (
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
)

const (
	ipcAPIs  = "admin:1.0 debug:1.0 ftm:1.0 net:1.0 personal:1.0 rpc:1.0 sfc:1.0 txpool:1.0 web3:1.0"
	httpAPIs = "ftm:1.0 rpc:1.0 sfc:1.0 web3:1.0"
)

// Tests that a node embedded within a console can be started up properly and
// then terminated by closing the input stream.
func TestConsoleWelcome(t *testing.T) {
	// Start a lachesis console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"console")

	// Gather all the infos the welcome message needs to contain
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

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
}

// Tests that a console can be attached to a running node via various means.
func TestIPCAttachWelcome(t *testing.T) {
	// Configure the instance for IPC attachement
	var ipc string
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\lachesis.ipc`
	} else {
		ws := tmpdir(t)
		defer os.RemoveAll(ws)
		ipc = filepath.Join(ws, "lachesis.ipc")
	}
	cli := exec(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--ipcpath", ipc)

	waitForEndpoint(t, ipc, 10*time.Second)
	testAttachWelcome(t, cli, "ipc:"+ipc, ipcAPIs)

	cli.Kill()
	cli.WaitExit()
}

func TestHTTPAttachWelcome(t *testing.T) {
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P
	cli := exec(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--rpc", "--rpcport", port)

	endpoint := "http://127.0.0.1:" + port
	waitForEndpoint(t, endpoint, 10*time.Second)
	testAttachWelcome(t, cli, "http://localhost:"+port, httpAPIs)

	cli.Kill()
	cli.WaitExit()
}

func TestWSAttachWelcome(t *testing.T) {
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P

	cli := exec(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--ws", "--wsport", port)

	endpoint := "ws://127.0.0.1:" + port
	waitForEndpoint(t, endpoint, 10*time.Second)
	testAttachWelcome(t, cli, "ws://localhost:"+port, httpAPIs)

	cli.Kill()
	cli.WaitExit()
}

func testAttachWelcome(t *testing.T, cli *testcli, endpoint, apis string) {
	// Attach to a running lachesis node and terminate immediately
	attach := exec(t, "attach", endpoint)

	// Gather all the infos the welcome message needs to contain
	attach.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	attach.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	attach.SetTemplateFunc("gover", runtime.Version)
	attach.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	attach.SetTemplateFunc("niltime", genesisStart)
	attach.SetTemplateFunc("ipc", func() bool { return strings.HasPrefix(endpoint, "ipc") })
	attach.SetTemplateFunc("datadir", func() string { return cli.Datadir })
	attach.SetTemplateFunc("coinbase", func() string { return cli.Coinbase })
	attach.SetTemplateFunc("apis", func() string { return apis })

	// Verify the actual welcome message to the required template
	attach.Expect(`
Welcome to the Lachesis JavaScript console!

instance: go-lachesis/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{coinbase}}
at block: 0 ({{niltime}}){{if ipc}}
 datadir: {{datadir}}{{end}}
 modules: {{apis}}

> {{.InputLine "exit" }}
`)
	attach.ExpectExit()
}

// trulyRandInt generates a crypto random integer used by the console tests to
// not clash network ports with other tests running cocurrently.
func trulyRandInt(lo, hi int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo)))
	return int(num.Int64()) + lo
}

func genesisStart() string {
	g := genesis.MainGenesis()
	s := g.Time.Unix()
	return time.Unix(s, 0).Format(time.RFC1123)
}
