package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/pkg/reexec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/cmd/cmdtest"
)

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "lachesis-test")
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
	addr := common.Address{} // TODO: addr = emitter coinbase
	tt.Coinbase = strings.ToLower(addr.String())
}

func init() {
	// Run the app if we've been exec'd as "lachesis-test" in exec().
	reexec.Register("lachesis-test", func() {
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

	// Boot "lachesis". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("lachesis-test", args...)

	// Read the generated key
	tt.readConfig()

	return tt
}

// waitForEndpoint attempts to connect to an RPC endpoint until it succeeds.
func waitForEndpoint(t *testing.T, endpoint string, timeout time.Duration) {
	probe := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c, err := rpc.DialContext(ctx, endpoint)
		if c != nil {
			_, err = c.SupportedModules()
			c.Close()
		}
		return err == nil
	}

	start := time.Now()
	for {
		if probe() {
			return
		}
		if time.Since(start) > timeout {
			t.Fatal("endpoint", endpoint, "did not open within", timeout)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
