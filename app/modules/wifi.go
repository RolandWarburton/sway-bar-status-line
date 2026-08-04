package modules

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var probeTargets = []string{"1.1.1.1:53", "8.8.8.8:53"}

const probeTimeout = 3 * time.Second

type Wifi struct {
	Module
	mu     sync.RWMutex
	device string // empty string when no wireless interface is up
	online bool   // associated interface can actually reach the internet
	polled bool
}

func (m *Wifi) Init() {
	m.Enabled = true
	go m.poll(10 * time.Second)
}

func (m *Wifi) poll(interval time.Duration) {
	for {
		device := activeWifiDevice()
		online := device != "" && hasConnectivity()

		m.mu.Lock()
		m.device, m.online, m.polled = device, online, true
		m.mu.Unlock()
		time.Sleep(interval)
	}
}

// Run reports one of four states:
//
//	┌──────────────────────────────────┬───────────────┐
//	│            Condition             │    Output     │
//	├──────────────────────────────────┼───────────────┤
//	│ Before first poll                │ Loading...    │
//	├──────────────────────────────────┼───────────────┤
//	│ No wireless iface up             │ WIFI DOWN     │
//	├──────────────────────────────────┼───────────────┤
//	│ Associated, no route off-network │ wlp4s0 NO NET │
//	├──────────────────────────────────┼───────────────┤
//	│ Associated + reachable           │ wlp4s0 UP     │
//	└──────────────────────────────────┴───────────────┘

func (m *Wifi) Run() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.polled {
		return "Loading..."
	}
	if m.device == "" {
		return "WIFI DOWN"
	}
	if !m.online {
		return fmt.Sprintf("%s NO NET", m.device)
	}
	return fmt.Sprintf("%s UP", m.device)
}

// returns true when TCP handshake can complete against any probe target
func hasConnectivity() bool {
	for _, target := range probeTargets {
		conn, err := net.DialTimeout("tcp4", target, probeTimeout)
		if err != nil {
			continue
		}
		conn.Close()
		return true
	}
	return false
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
