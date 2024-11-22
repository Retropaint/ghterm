package cmd

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CodeViewer string `toml:"code_viewer"`
	CodeEditor string `toml:"code_editor"`
}

func defaultConfig() Config {
	return Config{
		CodeViewer: "",
		CodeEditor: "",
	}
}

func loadCfg(cfg *Config) error {
	path, err := os.UserConfigDir()
	if err != nil {
		path = "."
	}

	path = filepath.Join(path, "ghterm", "config.toml")
	f, err := os.Open(path)

	*cfg = defaultConfig()
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}
	defer f.Close()

	_, err = toml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}
