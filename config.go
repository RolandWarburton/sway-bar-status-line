package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)
func defaultConfig() *Config {
	defaultSecrets := &Secrets{
		PTVDEVID: "",
		PTVKEY:   "",
	}

	defaultModules := &Modules{
		TIME:    true,
		PTV:     false,
		WIFI:    false,
		BATTERY: false,
	}

	defaultConfig := &Config{
		Secrets: *defaultSecrets,
		Modules: *defaultModules,
	}

	return defaultConfig
}

func getConfig() *Config {
	defaultConfig := defaultConfig()

	f, err := os.ReadFile("./config.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read config")
		return defaultConfig
	}

	var config Config
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal config")
		return defaultConfig
	}

	return &config
}
