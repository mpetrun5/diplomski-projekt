package chain

import (
	"fmt"

	"github.com/mpetrun5/diplomski-projekt/flags"
	"github.com/spf13/viper"
)

type GeneralChainConfig struct {
	Name           string `mapstructure:"name"`
	Id             *uint8 `mapstructure:"id"`
	Endpoint       string `mapstructure:"endpoint"`
	BlockstorePath string
	FreshStart     bool
	LatestBlock    bool
}

func (c *GeneralChainConfig) Validate() error {
	if c.Id == nil {
		return fmt.Errorf("required field domain.Id empty for chain %v", c.Id)
	}
	if c.Endpoint == "" {
		return fmt.Errorf("required field chain.Endpoint empty for chain %v", *c.Id)
	}
	if c.Name == "" {
		return fmt.Errorf("required field chain.Name empty for chain %v", *c.Id)
	}
	return nil
}

func (c *GeneralChainConfig) ParseFlags() {
	c.BlockstorePath = viper.GetString(flags.BlockstoreFlagName)
	c.FreshStart = viper.GetBool(flags.FreshStartFlagName)
	c.LatestBlock = viper.GetBool(flags.LatestBlockFlagName)
}
