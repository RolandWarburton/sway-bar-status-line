package modules

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const cpuPollInterval = 5 * time.Second

type Cpu struct {
	Module
	mu     sync.RWMutex
	usage  uint64 // percent, already clamped to 0-99
	polled bool
}

func (m *Cpu) Init() {
	m.Enabled = true
	go m.poll(cpuPollInterval)
}

// averages CPU utilization over the interval
// so each render tick is an average of the interval
func (m *Cpu) poll(interval time.Duration) {
	prevIdle, prevTotal, err := readCpuSample()
	// if we can read a sample without error
	seeded := err == nil

	for {
		time.Sleep(interval)

		idle, total, err := readCpuSample()
		if err != nil {
			continue
		}
		if !seeded {
			prevIdle, prevTotal, seeded = idle, total, true
			continue
		}

		deltaIdle, deltaTotal := idle-prevIdle, total-prevTotal
		prevIdle, prevTotal = idle, total
		if deltaTotal == 0 {
			continue
		}

		// clamp to two digits so the module keeps a constant width
		usage := min(100*(deltaTotal-deltaIdle)/deltaTotal, 99)

		m.mu.Lock()
		m.usage, m.polled = usage, true
		m.mu.Unlock()
	}
}

func (m *Cpu) Run() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.polled {
		return "CPU ..."
	}
	return fmt.Sprintf("CPU %02d%%", m.usage)
}

// Returns the cumulative idle and total jiffy counts
// Idle counts both idle and iowait time.
func readCpuSample() (idle uint64, total uint64, err error) {
	f, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}

	line, _, _ := strings.Cut(string(f), "\n")
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, fmt.Errorf("unexpected /proc/stat format")
	}

	for i, field := range fields[1:] {
		v, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		total += v
		// fields[1:] is user nice system idle iowait ...
		// cpu  2905331  8842   792604  48208842  12775   0    24184    0     0      0
		//      user     nice   system  idle      iowait  irq  softirq  steal guest  guest_nice
		if i == 3 || i == 4 {
			idle += v
		}
	}

	return idle, total, nil
}
