package types

type Secrets struct {
	PTVDEVID string `yaml:"PTV_DEVID"`
	PTVKEY   string `yaml:"PTV_KEY"`
}

type Modules struct {
	TIME    ModuleTime    `yaml:"TIME"`
	PTV     ModulePtv     `yaml:"PTV"`
	WIFI    ModuleWifi    `yaml:"WIFI"`
	BATTERY ModuleBattery `yaml:"BATTERY"`
}

type ModuleTime struct {
	Enabled bool
}

type ModulePtv struct {
	Enabled       bool   `yaml:"enabled"`
	RouteName     string `yaml:"routeName"`
	StopName      string `yaml:"stopName"`
	DirectionName string `yaml:"directionName"`
}

type ModuleWifi struct {
	Enabled bool
}

type ModuleBattery struct {
	Enabled bool
}

type Config struct {
	Secrets Secrets `yaml:"secrets"`
	Modules Modules `yaml:"modules"`
}
