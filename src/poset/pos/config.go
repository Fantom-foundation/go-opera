package pos

// Config for a PoS
type Config struct {
	TotalSupply uint64 `mapstructure:"total-supply"`
}

// NewConfig creates a new PoS config
func NewConfig(totalSupply uint64) *Config {
	return &Config{
		TotalSupply: totalSupply,
	}
}

// DefaultConfig sets the default config for a PoS
func DefaultConfig() *Config {
	return &Config{
		TotalSupply: 1000000000000000,
	}
}
