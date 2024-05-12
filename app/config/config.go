package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	logger "github.com/rolandwarburton/sway-status-line/app/logger"
	types "github.com/rolandwarburton/sway-status-line/app/types"
	"gopkg.in/yaml.v3"
)

func defaultConfig() *types.Config {
	defaultSecrets := &types.Secrets{
		PTVDEVID: "",
		PTVKEY:   "",
	}

	defaultModules := &types.Modules{
		TIME: types.ModuleTime{Enabled: true},
		PTV: types.ModulePtv{
			Enabled:       false,
			RouteName:     "Belgrave",
			StopName:      "Southern Cross",
			DirectionName: "Belgrave",
		},
		WIFI:    types.ModuleWifi{Enabled: false},
		BATTERY: types.ModuleBattery{Enabled: false},
	}

	defaultConfig := &types.Config{
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

func GetConfig() *types.Config {
	defaultConfig := defaultConfig()
	configPath, err := getConfigPath()
	if err != nil {
		logger.Alert(err.Error())
		return defaultConfig
	}

	f, err := os.ReadFile(*configPath)
	if err != nil {
		logger.Alert(err.Error())
		return defaultConfig
	}

	var config types.Config
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		logger.Alert(err.Error())
		return defaultConfig
	}

	return &config
}
