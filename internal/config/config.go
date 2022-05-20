package config

type Config struct {
	Log        LogConfig        `toml:"log"`
	Account    AccountConfig    `toml:"account"`
	Clients    ClientsConfig    `toml:"clients"`
	Strategies StrategiesConfig `toml:"strategies"`
}

type LogConfig struct {
	Level string `validate:"required"`
}

type AccountConfig struct {
	Number string `validate:"required"`
}

type ClientsConfig struct {
	TinkoffInvest TinkoffInvestConfig `toml:"tinkfoff_invest"`
}

type TinkoffInvestConfig struct {
	UseSandbox bool   `toml:"use_sandbox"`
	Address    string `toml:"address" validate:"required"`
	AppName    string `toml:"app_name" validate:"required"`
	Token      string `toml:"token" validate:"required"`
}

type StrategiesConfig struct {
	BullsAndBearsMonitoring BullsAndBearsMonitoringConfig `toml:"bulls_and_bears_monitoring"`
	SpreadMonitoring        SpreadMonitoringConfig        `toml:"spread_monitoring"`
}

type BullsAndBearsMonitoringConfig struct {
	Enabled            bool `toml:"enabled"`
	IgnoreInconsistent bool `toml:"ignore_inconsistent"`
	Instruments        []struct {
		FIGI             string  `toml:"figi" validate:"required"`
		Depth            int     `toml:"depth" validate:"required,oneof=[1 10 20 30 40 50]"`
		DominanceRatio   float64 `toml:"dominance_ratio" validate:"required,gt=1"`
		ProfitPercentage float64 `toml:"profit_percentage" validate:"required,gt=0,lte=1"`
	} `toml:"instruments" validate:"required,min=1"`
}

type SpreadMonitoringConfig struct {
	Enabled             bool    `toml:"enabled"`
	IgnoreInconsistent  bool    `toml:"ignore_inconsistent"`
	Depth               int     `toml:"depth" validate:"required,oneof=[1 10 20 30 40 50]"`
	MinSpreadPercentage float64 `toml:"min_spread_percentage" validate:"required,gt=0,lte=1"`
	MaxTools            int     `toml:"max_tools" validate:"required,gt=0"`
}
