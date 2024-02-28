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
	if err != nil {
		return "0%"
	} else {
		return fmt.Sprintf("%d%%", batteryCapacity)
	}
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
