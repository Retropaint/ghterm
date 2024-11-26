package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CodeViewer       string `toml:"code_viewer"`
	CodeViewerParams string `toml:"code_viewer_params"`
	CodeEditor       string `toml:"code_editor"`
	CodeEditorParams string `toml:"code_editor_params"`
}

var (
	Cfg Config
)

func defaultConfig() Config {
	return Config{
		CodeViewer:       "",
		CodeEditor:       "",
		CodeViewerParams: "",
		CodeEditorParams: "",
	}
}

func LoadCfg() error {
	path, err := os.UserConfigDir()
	if err != nil {
		path = "."
	}

	path = filepath.Join(path, "ghterm", "config.toml")
	f, err := os.Open(path)

	Cfg = defaultConfig()
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}
	defer f.Close()

	_, err = toml.NewDecoder(f).Decode(&Cfg)
	if err != nil {
		return err
	}

	return nil
}
