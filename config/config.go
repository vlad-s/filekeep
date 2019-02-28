package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

type web struct {
	Address string `yaml:"address"`
	Port    uint16 `yaml:"port"`
}

func (l web) String() string {
	return fmt.Sprintf("%s:%d", l.Address, l.Port)
}

// Config stores the config that the manager will use.
type Config struct {
	// Web defines the listening address and port.
	Web web `yaml:"web"`
	// Root is the root directory. Defaults to ".", the runtime dir.
	Root string `yaml:"root"`
	// Hidden is a slice of hidden files or directories.
	Hidden []string `yaml:"hidden"`
	// HiddenExts is a slice of hidden files by extension.
	HiddenExts []string `yaml:"hidden_extensions"`
	// Dotfiles will show files and directories starting with a dot.
	Dotfiles bool `yaml:"dotfiles"`
	// Debug will show additional debugging info. Verbose output, only switch if needed.
	Debug bool `yaml:"debug"`
}

// IsHidden returns whether a path is hidden or not. First checks if path is hidden by path,
// then by extension, then by being dotfile/dotdir.
func (c *Config) IsHidden(path ...string) bool {
	for _, p := range path {
		if c.IsHiddenPath(p) || c.IsHiddenExt(p) || c.IsHiddenDot(p) {
			return true
		}
	}
	return false
}

// IsHiddenPath returns whether a path is hidden by checking the Hidden slice in the config.
func (c *Config) IsHiddenPath(path string) bool {
	path = strings.ToLower(path)
	for _, v := range c.Hidden {
		v = strings.TrimRight(v, "/")
		v = strings.ToLower(v)
		base := filepath.Base(path)
		base = strings.ToLower(base)
		if v == path || v == base || strings.Index(path, v) == 0 || strings.Index(base, v) == 0 {
			return true
		}
	}
	return false
}

// IsHiddenExt returns whether a path is hidden by checking the HiddenExts slice in the config.
func (c *Config) IsHiddenExt(path string) bool {
	for _, v := range c.HiddenExts {
		if v == filepath.Ext(path) {
			return true
		}
	}
	return false
}

// IsHiddenDot returns whether a path is hidden by checking the Dotfiles bool in the config.
func (c *Config) IsHiddenDot(path string) bool {
	if c.Dotfiles || path == "" || path == "." {
		return false
	}
	if len(path) > 0 && path[0] == '.' {
		return true
	}
	if filepath.Base(path)[0] == '.' {
		return true
	}
	return false
}

var c = &Config{
	Web: web{
		Address: "localhost",
		Port:    8080,
	},
	Root:       ".",
	Hidden:     []string{"/etc/passwd"},
	HiddenExts: []string{".bak", ".DS_Store"},
	Dotfiles:   false,
}

// Get returns the current config.
func Get() *Config {
	return c
}

// Dump will dump the default config to disk.
func Dump() error {
	conf, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("couldn't marshal config to YAML: %s", err)
	}

	f, err := os.Create("config.example.yaml")
	if err != nil {
		return fmt.Errorf("couldn't open YAML config file for writing: %s", err)
	}

	if _, err := f.Write(conf); err != nil {
		return fmt.Errorf("couldn't dump default config to file: %s", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("couldn't close file after dumping config: %s", err)
	}

	return nil
}

// Load will read the config from disk.
func Load(configPath string) (*Config, error) {
	f, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config file from disk: %s", err)
	}

	readConf := new(Config)
	if err := yaml.Unmarshal(f, readConf); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal config from JSON: %s", err)
	}

	if readConf.Root == "" {
		readConf.Root = "."
	}

	if readConf.Web.Port == 0 {
		readConf.Web.Port = 8080
	}

	c = readConf
	return c, nil
}
