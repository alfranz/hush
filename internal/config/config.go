package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Checks map[string]Check `yaml:"checks" mapstructure:"checks"`
}

type Check struct {
	Cmd   string `yaml:"cmd" mapstructure:"cmd"`
	Label string `yaml:"label" mapstructure:"label"`
	Grep  string `yaml:"grep" mapstructure:"grep"`
	Tail  int    `yaml:"tail" mapstructure:"tail"`
	Head  int    `yaml:"head" mapstructure:"head"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".hush")
	v.SetConfigType("yaml")

	// Search current directory and parents
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		v.AddConfigPath(dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil, nil // No config file is fine
		}
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
