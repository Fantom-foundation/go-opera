package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/docker/pkg/reexec"

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
	Datadir   string
	Etherbase string
}

func init() {
	// Run the app if we've been exec'd as "glachesis-test" in runGeth.
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
	for i, arg := range args {
		switch {
		case arg == "-datadir" || arg == "--datadir":
			if i < len(args)-1 {
				tt.Datadir = args[i+1]
			}
		case arg == "-etherbase" || arg == "--etherbase":
			if i < len(args)-1 {
				tt.Etherbase = args[i+1]
			}
		}
	}
	if tt.Datadir == "" {
		tt.Datadir = tmpdir(t)
		tt.Cleanup = func() { os.RemoveAll(tt.Datadir) }
		args = append([]string{"-datadir", tt.Datadir}, args...)
		// Remove the temporary datadir if something fails below.
		defer func() {
			if t.Failed() {
				tt.Cleanup()
			}
		}()
	}

	// Boot "glachesis". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("glachesis-test", args...)

	return tt
}
