package main

import (
	"fmt"
	"os"
	"time"

	modules "github.com/rolandwarburton/sway-status-line/lib/modules"
)

func printStatus(timeModule *modules.Time, battery *modules.Battery, wifi *modules.Wifi) {
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

	fmt.Println(result)
}

func main() {
	timeModule := &modules.Time{}
	battery := &modules.Battery{}
	wifi := &modules.Wifi{}

	if os.Getenv("STATUS_SHOW_TIME") != "" {
		timeModule.Init()
	}

	if os.Getenv("STATUS_SHOW_BATTERY") != "" {
		battery.Init()
	}

	if os.Getenv("STATUS_SHOW_WIFI") != "" {
		wifi.Init()
	}

	go func() {
		for {
			printStatus(timeModule, battery, wifi)
			// Wait for 1 second before printing the next status
			time.Sleep(1 * time.Second)
		}

	}()
	select {}
}
