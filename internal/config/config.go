package config

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Configuration ...
type Configuration struct {
	Host          string `mapstructure:"HOST"`
	Port          int    `mapstructure:"PORT" validate:"required"`
	LogLevel      string `mapstructure:"LOG_LEVEL"`
	FakeAWSClient bool   `mapstructure:"FAKE_AWS_CLIENT"`
}

// GetConfig Gets configuration from environment variables
func GetConfig() (*Configuration, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	conf := &Configuration{
		Host: "",
		Port: 8080,
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
