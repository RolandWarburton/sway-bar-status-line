package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Wifi struct {
	Module
	mu     sync.RWMutex
	device string // "" when no wireless interface is up
	polled bool
}

func (m *Wifi) Init() {
	m.Enabled = true
	go m.poll(10 * time.Second)
}

func (m *Wifi) poll(interval time.Duration) {
	for {
		device := activeWifiDevice()
		m.mu.Lock()
		m.device, m.polled = device, true
		m.mu.Unlock()
		time.Sleep(interval)
	}
}

func (m *Wifi) Run() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.polled {
		return "Loading..."
	}
	if m.device != "" {
		return fmt.Sprintf("%s UP", m.device)
	}
	return "WIFI DOWN"
}

// activeWifiDevice returns the name of the first up wireless interface, or "".
// A wireless interface is identified by a "wireless" directory under its
// sysfs entry, and is considered up when its operstate reads "up".
func activeWifiDevice() string {
	const base = "/sys/class/net"
	entries, err := os.ReadDir(base)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if _, err := os.Stat(filepath.Join(base, e.Name(), "wireless")); err != nil {
			continue // not a wireless interface
		}
		state, err := os.ReadFile(filepath.Join(base, e.Name(), "operstate"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(state)) == "up" {
			return e.Name()
		}
	}
	return ""
}
