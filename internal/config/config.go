package config

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Configuration ...
type Configuration struct {
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
	DatabaseUser     string `mapstructure:"DATABASE_USER"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseHost     string `mapstructure:"DATABASE_HOST"`
	DatabasePort     string `mapstructure:"DATABASE_PORT"`

	// TODO: Add your own properties

	DataDogEnabled bool   `mapstructure:"DD_ENABLE"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`
}

// GetConfig Gets configuration from environment variables
func GetConfig() (*Configuration, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	conf := &Configuration{
		DatabaseHost:   "localhost",
		DatabasePort:   "5432",
		DataDogEnabled: false,
	}

	// Workaround for the viper to use environment variables without reading a config file
	// Viper #761: Unmarshal non-bound environment variables
	// https://github.com/spf13/viper/issues/761
	envKeysMap := &map[string]interface{}{}
	if err := mapstructure.Decode(conf, &envKeysMap); err != nil {
		return nil, fmt.Errorf("decoding conf: %w", err)
	}
	for k := range *envKeysMap {
		if err := viper.BindEnv(k); err != nil {
			return nil, fmt.Errorf("binding env key \"%s\": %w", k, err)
		}
	}
	// END (Workaround)

	if err := viper.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("unmarshaling: %w", err)
	}

	return conf, nil
}
