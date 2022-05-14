package config

type Config struct {
	Log     LogConfig     `toml:"log"`
	Clients ClientsConfig `toml:"clients"`
}

type LogConfig struct {
	Level string `validate:"required"`
}

type ClientsConfig struct {
	TinkoffInvest TinkoffInvestConfig `toml:"tinkfoff_invest"`
}

type TinkoffInvestConfig struct {
	Address string `toml:"address" validate:"required"`
	AppName string `toml:"app_name" validate:"required"`
	Token   string `toml:"token" validate:"required"`
}
