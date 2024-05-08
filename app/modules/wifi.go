package modules

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Wifi struct {
	Module
	networkState NetworkState
}

type NetworkState struct {
	Connected bool   `json:"state"`
	Device    string `json:"device"`
}

func (m *Wifi) Init() {
	m.Enabled = true
	m.networkState.Device = "..."
	go m.PollActiveWifiDevice(10 * time.Second)
}

func (m *Wifi) Run() string {
	var wifi string
	m.GetActiveWifiConnection()
	if m.networkState.Device != "" {
		wifi = fmt.Sprintf("%s UP", m.networkState.Device)
	} else {
		wifi = "WIFI DOWN"
	}
	return wifi
}

func (m *Wifi) GetActiveWifiConnection() error {
	connections, err := GetWifiDevices()
	found := false
	if err != nil {
		return err
	}
	for i := 0; i < len(connections); i += 1 {
		if connections[i].Connected {
			network := connections[i]
			m.networkState.Connected = network.Connected
			m.networkState.Device = network.Device
			found = true
		}
	}
	if found {
		return nil
	} else {
		return errors.New("no connected network was found")
	}
}

func GetWifiDevices() ([]NetworkState, error) {
	// get the devices
	cmd := exec.Command("nmcli", "-f", "GENERAL.state,GENERAL.device", "device", "show")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var networks []NetworkState
	for i := 0; i < len(lines)-1; i += 2 {
		var network NetworkState
		stateLine := lines[i]
		deviceLine := lines[i+1]
		if len(stateLine) == 0 || len(deviceLine) == 0 {
			continue
		}

		// STATE
		if stateLine[8:11] == "STA" {
			state := strings.TrimSpace(strings.Split(stateLine, "GENERAL.STATE:")[1])
			network.Connected = strings.Contains(state, "(connected)")
		}

		// DEVICE
		if deviceLine[8:11] == "DEV" {
			device := strings.TrimSpace(strings.Split(deviceLine, "GENERAL.DEVICE:")[1])
			network.Device = device
		}
		networks = append(networks, network)
	}

	return networks, nil
}

func (m *Wifi) PollActiveWifiDevice(timeout time.Duration) {
	for {
		m.GetActiveWifiConnection()
		time.Sleep(timeout)
	}
}
