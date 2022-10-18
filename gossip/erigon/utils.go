package erigon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"os/user"
)

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Lachesis")
			// linux
		default:
			return filepath.Join("/var/data", ".opera")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

// Fatalf formats a message to standard error and exits the program.
// The message is also printed to standard output if standard error
// is redirected to a different file.
func Fatalf(format string, args ...interface{}) {
	w := io.MultiWriter(os.Stdout, os.Stderr)

	outf, _ := os.Stdout.Stat()
	errf, _ := os.Stderr.Stat()
	if outf != nil && errf != nil && os.SameFile(outf, errf) {
		w = os.Stderr
	}

	fmt.Fprintf(w, "Fatal: "+format+"\n", args...)
	os.Exit(1)
}
