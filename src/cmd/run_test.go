package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/pkg/reexec"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-lachesis/src/cmd/cmdtest"
)

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "glachesis-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

type testcli struct {
	*cmdtest.TestCmd

	// template variables for expect
	Datadir  string
	Coinbase string
}

func (tt *testcli) readConfig() {
	cfg := defaultNodeConfig()
	cfg.DataDir = tt.Datadir
	addr := crypto.PubkeyToAddress(cfg.NodeKey().PublicKey)
	tt.Coinbase = strings.ToLower(addr.String())
}

func init() {
	// Run the app if we've been exec'd as "glachesis-test" in exec().
	reexec.Register("glachesis-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func TestMain(m *testing.M) {
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}

// exec cli with the given command line args. If the args don't set --datadir, the
// child g gets a temporary data directory.
func exec(t *testing.T, args ...string) *testcli {
	tt := &testcli{}
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)

	if len(args) < 1 || args[0] != "attach" {
		// make datadir
		for i, arg := range args {
			switch {
			case arg == "-datadir" || arg == "--datadir":
				if i < len(args)-1 {
					tt.Datadir = args[i+1]
				}
			}
		}
		if tt.Datadir == "" {
			tt.Datadir = tmpdir(t)
			args = append([]string{"-datadir", tt.Datadir}, args...)
		}

		// Remove the temporary datadir.
		tt.Cleanup = func() { os.RemoveAll(tt.Datadir) }
		defer func() {
			if t.Failed() {
				tt.Cleanup()
			}
		}()
	}

	// Boot "glachesis". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("glachesis-test", args...)

	// Read the generated key
	tt.readConfig()

	return tt
}
