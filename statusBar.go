package main

import (
	"fmt"
	"time"

	config "github.com/rolandwarburton/sway-status-line/app/config"
	modules "github.com/rolandwarburton/sway-status-line/app/modules"
)

func printStatus(timeModule *modules.Time, battery *modules.Battery, wifi *modules.Wifi, ptv *modules.PublicTransport) {
	var result string

	if timeModule.Enabled {
		result = timeModule.Run() + result
	}
	if battery.Enabled {
		result = battery.Run() + " | " + result
	}
	if wifi.Enabled {
		result = wifi.Run() + " | " + result
	}
	if ptv.Enabled {
		result = ptv.Run() + " | " + result
	}

	fmt.Println(result)
}

func main() {
	config := config.GetConfig()
	timeModule := &modules.Time{}
	battery := &modules.Battery{}
	wifi := &modules.Wifi{}
	ptv := &modules.PublicTransport{}

	if config.Modules.TIME {
		timeModule.Init()
	}

	if config.Modules.BATTERY {
		battery.Init()
	}

	if config.Modules.WIFI {
		wifi.Init()
	}

	if config.Modules.PTV {
		ptv.Init()
	}

	go func() {
		for {
			printStatus(timeModule, battery, wifi, ptv)
			// Wait for 1 second before printing the next status
			time.Sleep(1 * time.Second)
		}

	}()
	select {}
}
