package modules

import (
	"fmt"
	"time"

	ptv "github.com/rolandwarburton/ptv-status-line/pkg"
)

type PublicTransport struct {
	Module
	Departures []ptv.Departure
}

func (m *PublicTransport) Init() {
	m.Enabled = true
	go func() {
		for {
			departures, err := ptv.DeparturesAction("Lilydale", "Southern Cross", "Lilydale", 3, "Australia/Sydney")
			if err != nil {
				continue
			}
			m.Departures = departures
			time.Sleep(2 * time.Minute)
		}
	}()
}

func (m *PublicTransport) Run() string {
	if len(m.Departures) == 0 {
		return "no departures"
	}
	departure := m.Departures[0]

	// convert the departure time into a 12h short format (HH:MM)
	var location *time.Location
	layout := "02-01-2006 03:04 PM"
	departureTime, err := time.Parse(layout, departure.ScheduledDepartureUTC)
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
