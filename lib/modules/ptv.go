package modules

import (
	"errors"
	"fmt"
	"time"

	ptv "github.com/rolandwarburton/ptv-status-line/pkg"
)

type PublicTransport struct {
	Module
	Departures    []ptv.Departure
	NextDeparture time.Time
	LastPoll      time.Time
	nextPoll      time.Time
}

func parseStringToTime(timeString string) (time.Time, error) {
	var departureTime time.Time
	layout := "02-01-2006 03:04 PM"
	departureTime, err := time.Parse(layout, timeString)
	if err != nil {
		fmt.Println(err)
		return time.Time{}, errors.New("error parsing time")
	}
	return departureTime, nil
}

func parseTimeToLocaleString(o time.Time) string {
	// convert the departure time into a 12h short format (HH:MM)
	location, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		return ""
	}

	secUntilDeparture := time.Until(o).Seconds()
	var departureTimeString string
	if secUntilDeparture < 0 {
		departureTimeString = "waiting for next train"
		return fmt.Sprintf("%s %.0f", departureTimeString, secUntilDeparture)
	} else {
		departureTimeString = o.In(location).Format("03:04 PM")
		return fmt.Sprintf("Train in %.0fmin (%s)", secUntilDeparture, departureTimeString)
	}
}

func (m *PublicTransport) calcPollSleepTime() time.Duration {

	// set the sleep time based on the last train departure
	var sleepTime time.Duration

	nextDepartureTime, err := parseStringToTime(m.Departures[0].ScheduledDepartureUTC)

	if err != nil {
		return 5 * time.Minute
	}

	// default time to wait until next poll
	nextDepartureSeconds := time.Until(nextDepartureTime).Seconds()

	if nextDepartureSeconds <= 0 {
		sleepTime = 10 * time.Second
	} else {
		sleepTime = time.Duration(nextDepartureSeconds+10) * time.Second
	}
	return sleepTime
}

func (m *PublicTransport) poll() (sleep time.Duration) {
	fmt.Println("poll")
	departures, err := ptv.DeparturesAction("Belgrave", "Southern Cross", "Belgrave", 1, "Australia/Sydney")
	if err != nil || len(departures) == 0 {
		timeout := 5 * time.Minute
		m.nextPoll = time.Now().Add(timeout)
		return time.Duration(timeout)
	}
	m.Departures = departures
	m.LastPoll = time.Now()
	scheduledDepartureTime, _ := parseStringToTime(departures[0].ScheduledDepartureUTC)
	m.NextDeparture = scheduledDepartureTime

	sleepTime := m.calcPollSleepTime()
	m.nextPoll = time.Now().Add(sleepTime)
	return sleepTime
}

func (m *PublicTransport) startPoll() {
	go func() {
		for {
			sleepTime := m.poll()
			time.Sleep(sleepTime)
		}
	}()
}

func (m *PublicTransport) Init() {
	m.Enabled = true
	m.nextPoll = time.Now().Add(1 * time.Second)
	m.startPoll()
}

func (m *PublicTransport) Run() string {
	// test if there is anything to print
	if m.NextDeparture.Year() == 1 {
		nextDepartureSeconds := time.Until(m.nextPoll).Seconds()
		return fmt.Sprintf("polling in %.0f", nextDepartureSeconds)
	}

	var location *time.Location
	location, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		return "error"
	}

	timeUntilDeparture := time.Until(m.NextDeparture).Minutes()
	departureTimeString := m.NextDeparture.In(location).Format("03:04 PM")
	if timeUntilDeparture < 0 {
		departureTimeString = "waiting for next train"
		nextDepartureSeconds := time.Until(m.NextDeparture).Seconds()
		return fmt.Sprintf("%s %.0f", departureTimeString, nextDepartureSeconds)
	} else {
		return fmt.Sprintf("Train in %.0fmin (%s)", timeUntilDeparture, departureTimeString)
	}
}
