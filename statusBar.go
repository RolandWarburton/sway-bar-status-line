package main

import (
	"fmt"
	"os"
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
	config, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	timeModule := &modules.Time{}
	battery := &modules.Battery{}
	wifi := &modules.Wifi{}
	ptv := &modules.PublicTransport{}

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

	go func() {
		for {
			printStatus(timeModule, battery, wifi, ptv)
			// Wait for 1 second before printing the next status
			time.Sleep(1 * time.Second)
		}

	}()
	select {}
}
