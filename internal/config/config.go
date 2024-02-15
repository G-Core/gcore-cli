package config

import (
	"bytes"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Profile struct {
	Local   bool   `mapstructure:"local"    yaml:"local"`
	ApiURL  string `mapstructure:"url"      yaml:"url"`
	ApiKey  string `mapstructure:"apikey"   yaml:"apikey"`
	Project int    `mapstructure:"project"  yaml:"project"`
	Region  int    `mapstructure:"region"   yaml:"region"`
}

type Config struct {
	CurrentProfile string              `mapstructure:"current-profile" yaml:"current-profile"`
	Profiles       map[string]*Profile `mapstructure:"profiles"        yaml:"profiles"`
}

func NewDefault() Config {
	return Config{
		CurrentProfile: "default",
		Profiles: map[string]*Profile{
			"default": {
				Local:   false,
				ApiURL:  "https://api.gcore.com",
				ApiKey:  "",
				Project: 0,
				Region:  0,
			},
		},
	}
}

// Load tries to unmarshall viper into Config structure
func (c *Config) Load(v *viper.Viper) error {
	return v.Unmarshal(c)
}

// Save saves config in specified by viper config path in
// specified by viper name and format
func (c *Config) Save(v *viper.Viper) error {
	body, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err := v.ReadConfig(bytes.NewReader(body)); err != nil {
		return err
	}

	if err := v.WriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return v.SafeWriteConfig()
		}

		return err
	}

	return nil
}
