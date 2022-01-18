package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ChainConfigs []map[string]interface{}
}

type RawConfig struct {
	ChainConfigs []map[string]interface{} `mapstructure:"chains" json:"chains"`
}

// GetConfig reads config from file, validates it and parses
// it into config suitable for application
func GetConfig(path string) (Config, error) {
	rawConfig := RawConfig{}
	config := Config{}

	viper.SetConfigFile(path)
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&rawConfig)
	if err != nil {
		return config, err
	}

	for _, chain := range rawConfig.ChainConfigs {
		if chain["type"] == "" || chain["type"] == nil {
			return config, fmt.Errorf("Chain 'type' must be provided for every configured chain")
		}
	}

	config.ChainConfigs = rawConfig.ChainConfigs
	return config, nil
}
