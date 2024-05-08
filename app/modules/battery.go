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
	batteryState := m.readBatteryChargeState()
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

func (m *Battery) readBatteryChargeState() string {
	batteryFile, err := os.Open("/sys/class/power_supply/BAT0/status")
	if err != nil {
		return ""
	}
	defer batteryFile.Close()

	var batteryStatus string
	_, err = fmt.Fscanf(batteryFile, "%s", &batteryStatus)

	if err != nil {
		return ""
	}

	if batteryStatus == "Charging" {
		status := "^"
		return status
	}

	if batteryStatus == "Discharging" {
		status := ""
		return status
	}

	// if neither charging or discharging
	// the battery may be in an error state
	status := "!"
	return status
}
