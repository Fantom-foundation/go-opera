// +build multi

// This version will be built when tag MULTI is used in go build
//
package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runLachesis(cmd *cobra.Command, args []string) error {

	var configs [3]*CLIConfig

	for i := 0; i < 3; i++ {

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
		configs[i].Lachesis.DataDir += fmt.Sprintf("/%d", i)

		if i > 0 {
			go runSingleLachesis(configs[i])
		}

	}

	return runSingleLachesis(configs[0])
}
