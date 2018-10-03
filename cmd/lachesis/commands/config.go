package commands

import "github.com/andrecronje/lachesis/src/lachesis"

//CLIConfig contains configuration for the Run command
type CLIConfig struct {
	Lachesis     lachesis.LachesisConfig `mapstructure:",squash"`
	ProxyAddr  string              `mapstructure:"proxy-listen"`
	ClientAddr string              `mapstructure:"client-connect"`
	Inapp      bool                `mapstructure:"inapp"`
}

//NewDefaultCLIConfig creates a CLIConfig with default values
func NewDefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		Lachesis:     *lachesis.NewDefaultConfig(),
		ProxyAddr:  "127.0.0.1:1338",
		ClientAddr: "127.0.0.1:1339",
		Inapp:      false,
	}
}
