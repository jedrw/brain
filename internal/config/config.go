package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	// Flags
	AddressFlag        = "address"
	PortFlag           = "port"
	ContentDirFlag     = "content-dir"
	HostKeyPathFlag    = "host-key-path"
	AuthorizedKeysFlag = "authorized-keys"
	KeyPathFlag        = "key-path"

	// Defaults
	ContentDirDefault     = "./docs"
	HostKeyPathDefault    = "./id_ed25519"
	AuthorizedKeysDefault = ""
	AddressDefault        = ""
	PortDefault           = 2222
	UpdateTaskDefault     = "mkdocs build"
)

var (
	KeyPathDefault = filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")
)

type Config struct {
	Address        string   `yaml:"address"`
	Port           int      `yaml:"port"`
	HostKeyPath    string   `yaml:"hostKeyPath"`
	AuthorizedKeys []string `yaml:"authorizedKeys"`
	KeyPath        string   `yaml:"keyPath"`
	ContentDir     string   `yaml:"contentDir"`
	UpdateTasks    []string `yaml:"updateTasks"`
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

func (c *Config) setOverrides(flags *pflag.FlagSet) {
	if c.Address == "" || isFlagSet(AddressFlag) {
		c.Address, _ = flags.GetString(AddressFlag)
	}

	if c.Port == 0 || isFlagSet(PortFlag) {
		c.Port, _ = flags.GetInt(PortFlag)
	}

	if c.ContentDir == "" || isFlagSet(ContentDirFlag) {
		c.ContentDir, _ = flags.GetString(ContentDirFlag)
	}

	if c.HostKeyPath == "" || isFlagSet(HostKeyPathFlag) {
		c.HostKeyPath, _ = flags.GetString(HostKeyPathFlag)
	}

	if len(c.UpdateTasks) == 0 {
		c.UpdateTasks = append(c.UpdateTasks, UpdateTaskDefault)
	}

	if len(c.AuthorizedKeys) == 0 || isFlagSet(AuthorizedKeysFlag) {
		keysString, _ := flags.GetString(AuthorizedKeysFlag)
		keys := strings.SplitSeq(keysString, ",")
		for key := range keys {
			c.AuthorizedKeys = append(c.AuthorizedKeys, strings.TrimSpace(key))
		}
	}

	if c.KeyPath == "" || isFlagSet(KeyPathFlag) {
		c.KeyPath, _ = flags.GetString(KeyPathFlag)
	}
}

func New(configPath string, flags *pflag.FlagSet) (Config, error) {
	if configPath == "" {
		configPath = filepath.Join(xdg.ConfigHome, "brain", "config.yaml")
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return config, err
	}

	config.setOverrides(flags)

	return config, nil
}
