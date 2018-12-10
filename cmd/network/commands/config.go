package commands

import "github.com/Fantom-foundation/go-lachesis/src/lachesis"

//CLIConfig contains configuration for the Run command
type CLIConfig struct {
	Lachesis lachesis.LachesisConfig `mapstructure:",squash"`
	NbNodes  int                     `mapstructure:"nodes"`
	SendTxs  int                     `mapstructure:"send-txs"`
	Stdin    bool                    `mapstructure:"stdin"`
	Node     int                     `mapstructure:"node"`
}

//NewDefaultCLIConfig creates a CLIConfig with default values
func NewDefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		Lachesis: *lachesis.NewDefaultConfig(),
		NbNodes:  4,
		SendTxs:  0,
		Stdin:    false,
		Node:     0,
	}
}
