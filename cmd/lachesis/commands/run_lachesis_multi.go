// +build multi

// This version will be built when tag MULTI is used in go build
//
package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runLachesis(cmd *cobra.Command, args []string) error {

	var n uint = 3
	nValue := os.Getenv("n")
	if len(nValue) > 0 {
		n64, err := strconv.ParseUint(nValue, 10, 64)
		if err != nil {
			return err
		}
		n = uint(n64)
	}
	configs := make([]*CLIConfig, n)
	digits := len(strconv.FormatUint(uint64(n), 10))

	for i := uint(0); i < n; i++ {

		configs[i] = NewDefaultCLIConfig()

		if err := bindFlagsLoadViper(cmd, configs[i]); err != nil {
			return err
		}

		err := viper.Unmarshal(configs[i])
		if err != nil {
			return err
		}

		configs[i].Lachesis.BindAddr = fmt.Sprintf("127.0.0.1:%d", 12000 + i + 1)
		configs[i].Lachesis.ServiceAddr = fmt.Sprintf("127.0.0.1:%d", 8000 + i + 1)
		configs[i].ProxyAddr = fmt.Sprintf("127.0.0.1:%d", 9000 + i + 1)
		configs[i].Lachesis.DataDir += fmt.Sprintf("/%0*d", digits, i)

		if i > 0 {
			go runSingleLachesis(configs[i])
		}

	}

	return runSingleLachesis(configs[0])
}
