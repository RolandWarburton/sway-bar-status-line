package modules

import (
	"fmt"
	"math"
	"time"

	ptv "github.com/rolandwarburton/ptv-status-line/pkg"
)

type PublicTransport struct {
	Module
	Departures    []ptv.Departure
	NextDeparture time.Time
	Status        string
}

func (m *PublicTransport) Init() {
	m.Enabled = true
	m.Status = "loading..."
	go func() {
		for {
			departures, err := ptv.DeparturesAction("Lilydale", "Southern Cross", "Lilydale", 3, "Australia/Sydney")
			if err != nil {
				continue
			}
			m.Departures = departures

			// set the sleep time based on the last train departure
			sleepTime := 5 * time.Minute
			timeUntilDeparture := time.Until(m.NextDeparture).Minutes()
			if timeUntilDeparture < 0 {
				sleepTime = time.Duration(0)
			}
			if timeUntilDeparture > 5 {
				sleepTime = time.Duration(int(math.Ceil(timeUntilDeparture-1))) * time.Minute
			}

			time.Sleep(sleepTime)
		}
	}()
}

func (m *PublicTransport) Run() string {
	if len(m.Departures) == 0 {
		return m.Status
	}
	departure := m.Departures[0]

	// convert the departure time into a 12h short format (HH:MM)
	var location *time.Location
	layout := "02-01-2006 03:04 PM"
	departureTime, err := time.Parse(layout, departure.ScheduledDepartureUTC)
	m.NextDeparture = departureTime
	if err != nil {
		fmt.Println(err)
		return "error"
	}
	location, err = time.LoadLocation("Australia/Sydney")
	if err != nil {
		return "error"
	}

	timeUntilDeparture := time.Until(departureTime).Minutes()
	departureTimeString := departureTime.In(location).Format("03:04 PM")
	if timeUntilDeparture < 0 {
		departureTimeString = "waiting for next train"
	}
	return fmt.Sprintf("Train in %.0fmin (%s)", timeUntilDeparture, departureTimeString)
}
