package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Fantom-foundation/go-opera/cmd/opera/launcher"
)

func main() {
	// TODO erase after compatibility issues with go1.20 are fixed
	var majorVer int
	var minorVer int
	var other string
	n, err := fmt.Sscanf(runtime.Version(), "go%d.%d%s", &majorVer, &minorVer, &other)
	if n >= 2 && err == nil {
		if (majorVer*100 + minorVer) > 119 {
			panic(runtime.Version() + " is not supported, please downgrade your go compiler to go 1.19 or older")
		}
	}
	if err := launcher.Launch(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
