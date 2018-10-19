// +build !multi

// This version will be built when no tag MULTI is used in go build
//
package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runLachesis(cmd *cobra.Command, args []string) error {

	config := NewDefaultCLIConfig()

	if err := bindFlagsLoadViper(cmd, config); err != nil {
		return err
	}

	err := viper.Unmarshal(config)
	if err != nil {
		return err
	}

	return runSingleLachesis(config)
}
