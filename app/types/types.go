package types

type Modules struct {
	TIME    ModuleTime    `yaml:"TIME"`
	PTV     ModulePtv     `yaml:"PTV"`
	WIFI    ModuleWifi    `yaml:"WIFI"`
	BATTERY ModuleBattery `yaml:"BATTERY"`
	CPU     ModuleCpu     `yaml:"CPU"`
	MEMORY  ModuleMemory  `yaml:"MEMORY"`
	DISK    ModuleDisk    `yaml:"DISK"`
}

type ModuleTime struct {
	Enabled bool
}

type ModulePtv struct {
	PTVDEVID      string `yaml:"ptv_devid"`
	PTVKEY        string `yaml:"ptv_key"`
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

type ModuleCpu struct {
	Enabled bool `yaml:"enabled"`
}

type ModuleMemory struct {
	Enabled bool `yaml:"enabled"`
}

type ModuleDisk struct {
	Enabled bool `yaml:"enabled"`
	Partitions []DiskPartition `yaml:"partitions"`
}

type DiskPartition struct {
	Path string `yaml:"path"`
	Label string `yaml:"label"`
}

type Config struct {
	Modules Modules `yaml:"modules"`
}
