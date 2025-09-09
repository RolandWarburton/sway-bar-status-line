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

func GetConfig() (*types.Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		logger.Alert(fmt.Sprintf("failed to get the config path: %s", err.Error()))
		return nil, err
	}

	f, err := os.ReadFile(*configPath)
	if err != nil {
		logger.Alert(fmt.Sprintf("failed to read the config file: %s", err.Error()))
		return nil, err
	}

	var config types.Config
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		logger.Alert(fmt.Sprintf("failed to parse the config file: %s", err.Error()))
		return nil, err
	}

	return &config, nil
}
