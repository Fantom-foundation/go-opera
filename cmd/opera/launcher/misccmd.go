package launcher

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli/v2"

	"github.com/Fantom-foundation/go-opera/gossip"
)

var (
	versionCommand = &cli.Command{
		Action:    version,
		Name:      "version",
		Usage:     "Print version numbers",
		ArgsUsage: " ",
		Category:  "MISCELLANEOUS COMMANDS",
		Description: `
The output of this command is supposed to be machine-readable.
`,
	}

	licenseCommand = &cli.Command{
		Action:    license,
		Name:      "license",
		Usage:     "Display license information",
		ArgsUsage: " ",
		Category:  "MISCELLANEOUS COMMANDS",
	}
)

func version(ctx *cli.Context) error {
	fmt.Println(strings.Title(clientIdentifier))
	fmt.Println("Version:", params.VersionWithMeta())
	if gitCommit != "" {
		fmt.Println("Git Commit:", gitCommit)
	}
	if gitDate != "" {
		fmt.Println("Git Commit Date:", gitDate)
	}
	fmt.Println("Architecture:", runtime.GOARCH)
	fmt.Println("Protocol Versions:", []uint{gossip.ProtocolVersion})
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Operating System:", runtime.GOOS)
	fmt.Printf("GOPATH=%s\n", os.Getenv("GOPATH"))
	fmt.Printf("GOROOT=%s\n", runtime.GOROOT())
	return nil
}

func license(_ *cli.Context) error {
	// TODO: license text
	fmt.Println(``)
	return nil
}
