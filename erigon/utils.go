package erigon

import (
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

func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Lachesis")
			// linux
		default:
			return filepath.Join(home, ".opera")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}
