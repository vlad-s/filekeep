package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type listen struct {
	Addr string `json:"addr"`
	Port uint16 `json:"port"`
}

// Config stores the config that the manager will use.
type Config struct {
	// Listen defines the listening address and port.
	Listen listen `json:"listen"`
	// Root is the root directory. Defaults to ".", the runtime dir.
	Root string `json:"root"`
	// Hide is a slice of hidden files or directories.
	Hide []string `json:"hide"`
	// HideExt is a slice of hidden files by extension.
	HideExt []string `json:"hide_extensions"`
	// HideDots will hide (hidden) files and directories starting with a dot.
	HideDots bool `json:"hide_dots"`
	// Debug will show additional debugging info. Verbose output, only switch if needed.
	Debug bool `json:"debug"`
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

// IsHiddenPath returns whether a path is hidden by checking the Hide slice in the config.
func (c *Config) IsHiddenPath(path string) bool {
	path = strings.ToLower(path)
	for _, v := range c.Hide {
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

// IsHiddenExt returns whether a path is hidden by checking the HideExt slice in the config.
func (c *Config) IsHiddenExt(path string) bool {
	for _, v := range c.HideExt {
		if v == filepath.Ext(path) {
			return true
		}
	}
	return false
}

// IsHiddenDot returns whether a path is hidden by checking the HideDots bool in the config.
func (c *Config) IsHiddenDot(path string) bool {
	if !c.HideDots || path == "" || path == "." {
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
	Listen: listen{
		Addr: "",
		Port: 8080,
	},
	Root:     ".",
	Hide:     []string{},
	HideExt:  []string{".bak", ".DS_Store"},
	HideDots: true,
}

// Get returns the current config.
func Get() *Config {
	return c
}

// Dump will dump the default config to disk to "config.json".
func Dump() error {
	conf, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, "couldn't marshal config to JSON")
	}

	f, err := os.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "couldn't open JSON config file for write")
	}

	if _, err := f.Write(conf); err != nil {
		return errors.Wrap(err, "couldn't dump default config to file")
	}

	return errors.Wrap(f.Close(), "couldn't close file after dumping")
}

// Load will read the config from disk from "config.json".
func Load() error {
	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		return errors.Wrap(err, "couldn't read config file from disk")
	}

	readConf := new(Config)
	if err := json.Unmarshal(f, readConf); err != nil {
		return errors.Wrap(err, "couldn't unmarshal config from JSON")
	}

	if readConf.Root == "" {
		readConf.Root = "."
	}

	if readConf.Listen.Port == 0 {
		readConf.Listen.Port = 8080
	}

	c = readConf
	return nil
}
