package modules

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Memory struct {
	Module
}

func (m *Memory) Init() {
	m.Enabled = true
}

func (m *Memory) Run() string {
	total, available, err := readMemInfo()
	if err != nil || total == 0 {
		return "RAM ERR"
	}

	// excludes cache and buffers the kernel would hand back on demand
	used := total - available

	return fmt.Sprintf("RAM %02d%%", min(100*used/total, 99))
}

// readMemInfo returns MemTotal and MemAvailable from /proc/meminfo, in kB.
func readMemInfo() (total uint64, available uint64, err error) {
	f, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}

	for line := range strings.SplitSeq(string(f), "\n") {
		key, rest, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		if key != "MemTotal" && key != "MemAvailable" {
			continue
		}

		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		v, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return 0, 0, err
		}

		if key == "MemTotal" {
			total = v
		} else {
			available = v
		}
		if total != 0 && available != 0 {
			break
		}
	}

	if total == 0 {
		return 0, 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
	}
	return total, available, nil
}
