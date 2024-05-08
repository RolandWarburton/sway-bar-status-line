package main

import (
	"errors"
	"fmt"
	"os"
	"path"

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

func getConfigPath() (*string, error) {
	var home, err = os.UserHomeDir()
	if err != nil {
		// fall back on the default config
		return nil, errors.New("failed to get users home")
	}

	configLocation, cfgPathSet := os.LookupEnv("SWAYBAR_CONFIG_LOCATION")

	if !cfgPathSet {
		// default to ~/.config/swaybar/config.yml
		p := path.Join(home, ".config/swaybar/config.yml")
		return &p, nil
	}

	if _, err := os.Stat(configLocation); err != nil {
		// fall back on the default config
		return nil, errors.New("SWAYBAR_CONFIG_LOCATION does not exist")
	}

	return &configLocation, nil
}

func getConfig() *Config {
	defaultConfig := defaultConfig()
	configPath, err := getConfigPath()
	if err != nil {
		return defaultConfig
	}

	f, err := os.ReadFile(*configPath)
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
