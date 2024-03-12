package modules

import (
	"fmt"
	"os"
)

type Battery struct {
	Module
}

func (m *Battery) Init() {
	m.Enabled = true
}

func (m *Battery) Run() string {
	batteryCapacity, err := m.readBatteryCapacity()
	batteryState, _ := m.readBatteryChargeState()
	if err != nil {
		return "0%"
	}

	return fmt.Sprintf("%s%d%%", batteryState, batteryCapacity)
}

func (m *Battery) readBatteryCapacity() (int, error) {
	batteryFile, err := os.Open("/sys/class/power_supply/BAT0/capacity")
	if err != nil {
		return 0, err
	}
	defer batteryFile.Close()

	var batteryCapacity int
	_, err = fmt.Fscanf(batteryFile, "%d", &batteryCapacity)
	if err != nil {
		return 0, err
	}

	return batteryCapacity, nil
}

func (m *Battery) readBatteryChargeState() (string, error) {
	batteryFile, err := os.Open("/sys/class/power_supply/BAT0/status")
	if err != nil {
		return "", err
	}
	defer batteryFile.Close()

	var batteryStatus string
	_, err = fmt.Fscanf(batteryFile, "%s", &batteryStatus)

	if err != nil {
		return "", err
	}

	if batteryStatus == "Charging" {
		status := "^"
		return status, nil
	}

	if batteryStatus == "Discharging" {
		status := ""
		return status, nil
	}

	// if neither charging or discharging
	// the battery may be in an error state
	status := "!"
	return status, nil
}
