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
	LastPoll      time.Time
	nextPoll      time.Time
	Status        string
}

func (m *PublicTransport) poll() {
	// avoid sending more than one request at a time
	if m.Status == "polling" {
		return
	}
	m.Status = "polling"
	departures, err := ptv.DeparturesAction("Lilydale", "Southern Cross", "Lilydale", 3, "Australia/Sydney")
	if err != nil {
		m.Status = "error"
		return
	}
	m.Departures = departures
	m.LastPoll = time.Now()
	m.Status = "done"
}

func (m *PublicTransport) Init() {
	m.Enabled = true
	m.Status = "loading..."
	go func() {
		for {
			m.poll()
			// set the sleep time based on the last train departure
			sleepTime := 5 * time.Minute
			var seccondsUntilDeparture float64

			// if we are loading then wait 5s to throttle retries
			if m.Status == "loading..." {
				seccondsUntilDeparture = 5
			} else {
				seccondsUntilDeparture = time.Until(m.NextDeparture).Seconds()
			}

			// if the train left
			if seccondsUntilDeparture < 0 {
				sleepTime = time.Duration(0)
			}

			// if the train is going to leave in a long time slow down
			if seccondsUntilDeparture > 600 {
				sleepTime = time.Duration(int(math.Ceil(seccondsUntilDeparture/2))) * time.Minute
			}
			m.nextPoll = time.Now().Add(sleepTime)
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
		m.poll()
	}

	return fmt.Sprintf("Train in %.0fmin (%s)", timeUntilDeparture, departureTimeString)
}
