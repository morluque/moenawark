package config

import (
	"github.com/BurntSushi/toml"
	"os"
)

type Config struct {
	DBPath     string `toml:"db_path"`
	AdminLogin string `toml:"admin_login"`
}

func Parse(path string) (*Config, error) {
	var conf Config

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
