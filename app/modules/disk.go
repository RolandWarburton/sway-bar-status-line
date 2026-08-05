package modules

import (
	"fmt"
	"strings"
	"syscall"

	types "github.com/rolandwarburton/sway-status-line/app/types"
)

const diskPartitionSeparator = " "

type Disk struct {
	Module
	partitions []types.DiskPartition
}

func (m *Disk) Init(config types.ModuleDisk) {
	m.Enabled = true
	m.partitions = config.Partitions

	// Fall back to the root filesystem rather than rendering nothing
	if len(m.partitions) == 0 {
		m.partitions = []types.DiskPartition{{Path: "/"}}
	}

	// Label is optional
	for i, p := range m.partitions {
		if p.Label == "" {
			m.partitions[i].Label = p.Path
		}
	}
}

func (m *Disk) Run() string {
	parts := make([]string, 0, len(m.partitions))
	for _, p := range m.partitions {
		usage, err := diskUsagePercent(p.Path)
		if err != nil {
			parts = append(parts, fmt.Sprintf("%s ERR", p.Label))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %02d%%", p.Label, usage))
	}
	return strings.Join(parts, diskPartitionSeparator)
}

func diskUsagePercent(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}

	used := stat.Blocks - stat.Bfree
	total := used + stat.Bavail
	if total == 0 {
		return 0, fmt.Errorf("%s reports zero usable blocks", path)
	}

	// clamp to two digits (unless 100% used)
	return min(100*used/total, 100), nil
}
