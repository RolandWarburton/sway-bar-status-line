package types

type Secrets struct {
	PTVDEVID string `yaml:"PTV_DEVID"`
	PTVKEY   string `yaml:"PTV_KEY"`
}

type Modules struct {
	TIME    bool `yaml:"TIME"`
	PTV     bool `yaml:"PTV"`
	WIFI    bool `yaml:"WIFI"`
	BATTERY bool `yaml:"BATTERY"`
}

type Config struct {
	Secrets Secrets `yaml:"secrets"`
	Modules Modules `yaml:"modules"`
}
