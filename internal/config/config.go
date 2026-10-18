package config

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Defaults Defaults         `yaml:"defaults" mapstructure:"defaults"`
	Checks   map[string]Check `yaml:"checks" mapstructure:"checks"`
	// CheckOrder lists the names in Checks in the order of the config file.
	CheckOrder []string `yaml:"-" mapstructure:"-"`
}

type Defaults struct {
	Tail        int    `yaml:"tail" mapstructure:"tail"`
	Head        int    `yaml:"head" mapstructure:"head"`
	Grep        string `yaml:"grep" mapstructure:"grep"`
	WarnPattern string `yaml:"warn-pattern" mapstructure:"warn-pattern"`
	WarnTail    int    `yaml:"warn-tail" mapstructure:"warn-tail"`
	Continue    bool   `yaml:"continue" mapstructure:"continue"`
}

type Check struct {
	Cmd         string `yaml:"cmd" mapstructure:"cmd"`
	Label       string `yaml:"label" mapstructure:"label"`
	Grep        string `yaml:"grep" mapstructure:"grep"`
	WarnPattern string `yaml:"warn-pattern" mapstructure:"warn-pattern"`
	WarnTail    int    `yaml:"warn-tail" mapstructure:"warn-tail"`
	Tail        int    `yaml:"tail" mapstructure:"tail"`
	Head        int    `yaml:"head" mapstructure:"head"`
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
	order, err := checkOrder(v.ConfigFileUsed(), cfg.Checks)
	if err != nil {
		return nil, err
	}
	cfg.CheckOrder = order
	return &cfg, nil
}

// checkOrder returns the check names in the order of the config file.
// Viper loses this order, so it reads the file again. Viper also lowercases
// keys, so the names are lowercased to match the keys in checks.
func checkOrder(path string, checks map[string]Check) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Checks yaml.Node `yaml:"checks"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	var order []string
	if doc.Checks.Kind == yaml.MappingNode {
		for i := 0; i < len(doc.Checks.Content); i += 2 {
			name := strings.ToLower(doc.Checks.Content[i].Value)
			if _, ok := checks[name]; ok && !slices.Contains(order, name) {
				order = append(order, name)
			}
		}
	}
	// Add any check the file order missed, so that no check is skipped.
	for _, name := range slices.Sorted(maps.Keys(checks)) {
		if !slices.Contains(order, name) {
			order = append(order, name)
		}
	}
	return order, nil
}
