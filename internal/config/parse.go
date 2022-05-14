package config

import "github.com/BurntSushi/toml"

func Parse(filename string) (cfg Config, err error) {
	_, err = toml.DecodeFile(filename, &cfg)
	return
}
