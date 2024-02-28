package main

import (
	"fmt"
	"os"
	"time"

	modules "github.com/rolandwarburton/sway-status-line/lib/modules"
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
	timeModule := &modules.Time{}
	battery := &modules.Battery{}
	wifi := &modules.Wifi{}
	ptv := &modules.PublicTransport{}

	if os.Getenv("STATUS_SHOW_TIME") != "" {
		timeModule.Init()
	}

	if os.Getenv("STATUS_SHOW_BATTERY") != "" {
		battery.Init()
	}

	if os.Getenv("STATUS_SHOW_WIFI") != "" {
		wifi.Init()
	}

	if os.Getenv("STATUS_SHOW_PTV") != "" {
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
