package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	config "github.com/rolandwarburton/sway-status-line/app/config"
	modules "github.com/rolandwarburton/sway-status-line/app/modules"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func printStatus(
	timeModule *modules.Time,
	battery *modules.Battery,
	wifi *modules.Wifi,
	ptv *modules.PublicTransport,
	cpu *modules.Cpu,
	memory *modules.Memory,
) {

	var parts []string

	if cpu.Enabled {
		parts = append(parts, cpu.Run())
	}
	if memory.Enabled {
		parts = append(parts, memory.Run())
	}
	if ptv.Enabled {
		parts = append(parts, ptv.Run())
	}
	if wifi.Enabled {
		parts = append(parts, wifi.Run())
	}
	if battery.Enabled {
		parts = append(parts, battery.Run())
	}
	if timeModule.Enabled {
		parts = append(parts, timeModule.Run())
	}

	fmt.Println(strings.Join(parts, " | "))
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(version)
		return
	}

	config, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	timeModule := &modules.Time{}
	battery := &modules.Battery{}
	wifi := &modules.Wifi{}
	ptv := &modules.PublicTransport{}
	cpu := &modules.Cpu{}
	memory := &modules.Memory{}

	if config.Modules.TIME.Enabled {
		timeModule.Init()
	}

	if config.Modules.BATTERY.Enabled {
		battery.Init()
	}

	if config.Modules.WIFI.Enabled {
		wifi.Init()
	}

	if config.Modules.PTV.Enabled {
		ptv.Init(config.Modules.PTV)
	}

	if config.Modules.CPU.Enabled {
		cpu.Init()
	}

	if config.Modules.MEMORY.Enabled {
		memory.Init()
	}
	go func() {
		for {
			printStatus(timeModule, battery, wifi, ptv, cpu, memory)
			// Wait for 1 second before printing the next status
			time.Sleep(1 * time.Second)
		}

	}()
	select {}
}
