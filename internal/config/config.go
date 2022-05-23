package config

type Config struct {
	Log        LogConfig        `toml:"log"`
	Metrics    MetricsConfig    `toml:"metrics"`
	Account    AccountConfig    `toml:"account"`
	Clients    ClientsConfig    `toml:"clients"`
	Strategies StrategiesConfig `toml:"strategies"`
}

type LogConfig struct {
	Level string `validate:"required"`
}

type MetricsConfig struct {
	Enabled bool   `toml:"enabled"`
	Addr    string `toml:"addr" validate:"hostname_port"`
}

type AccountConfig struct {
	Number  string `validate:"required"`
	Sandbox bool   `toml:"sandbox"`
}

type ClientsConfig struct {
	TinkoffInvest TinkoffInvestConfig `toml:"tinkfoff_invest"`
}

type TinkoffInvestConfig struct {
	Address string `toml:"address" validate:"required"`
	AppName string `toml:"app_name" validate:"required"`
	Token   string `toml:"token" validate:"required"`
}

type StrategiesConfig struct {
	BullsAndBearsMonitoring BullsAndBearsMonitoringConfig `toml:"bulls_and_bears_monitoring"`
	SpreadParasite          SpreadParasiteConfig          `toml:"spread_parasite"`
}

type BullsAndBearsMonitoringConfig struct {
	Enabled            bool `toml:"enabled"`
	IgnoreInconsistent bool `toml:"ignore_inconsistent"`
	Instruments        []struct {
		FIGI             string  `toml:"figi" validate:"required"`
		Depth            int     `toml:"depth" validate:"required,oneof=[1 10 20 30 40 50]"`
		DominanceRatio   float64 `toml:"dominance_ratio" validate:"required,gt=1"`
		ProfitPercentage float64 `toml:"profit_percentage" validate:"required,gt=0,lte=1"`
	} `toml:"instruments" validate:"required,dive,min=1"`
}

type SpreadParasiteConfig struct {
	Enabled             bool     `toml:"enabled"`
	IgnoreInconsistent  bool     `toml:"ignore_inconsistent"`
	MinSpreadPercentage float64  `toml:"min_spread_percentage" validate:"required,gt=0,lte=1"`
	Figis               []string `toml:"figis"`
}
