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
	isPolling     bool
}

func (m *PublicTransport) poll() {
	// avoid sending more than one request at a time
	defer func() {
		m.isPolling = false
	}()
	if m.isPolling {
		return
	}
	m.isPolling = true
	departures, err := ptv.DeparturesAction("Lilydale", "Southern Cross", "Lilydale", 3, "Australia/Sydney")
	if err != nil {
		return
	}
	m.Departures = departures
	m.LastPoll = time.Now()
	m.LastPoll = time.Now()
	m.isPolling = false
}

func (m *PublicTransport) Init() {
	m.Enabled = true
	m.nextPoll = time.Now()
	go func() {
		for {
			m.poll()
			// sleep 5min
			sleepTime := 5 * time.Minute
			m.nextPoll = time.Now().Add(sleepTime)
			time.Sleep(sleepTime)
		}
	}()
}

type DepartureTimingInformation struct {
	DepartureTime         time.Time
	MinutesUntilDeparture float64
	Location              *time.Location
}

func GetDepartureTimingInformation(departure ptv.Departure) (*DepartureTimingInformation, error) {
	var location *time.Location
	layout := "02-01-2006 03:04 PM"
	departureTime, err := time.Parse(layout, departure.ScheduledDepartureUTC)
	if err != nil {
		return nil, err
	}
	location, err = time.LoadLocation("Australia/Sydney")
	if err != nil {
		return nil, err
	}
	minutesUntilDeparture := time.Until(departureTime).Minutes()
	return &DepartureTimingInformation{
		DepartureTime:         departureTime,
		MinutesUntilDeparture: minutesUntilDeparture,
		Location:              location,
	}, nil
}

func (m *PublicTransport) Run() string {
	if len(m.Departures) == 0 {
		secondsUntilNextPoll := math.Abs(time.Until(m.nextPoll).Seconds())
		isPollingStr := map[bool]string{true: " (polling)", false: ""}[m.isPolling]
		return fmt.Sprintf("No data available%s (%.0f)", isPollingStr, secondsUntilNextPoll)
	}

	var result string
	for i, departure := range m.Departures {
		info, err := GetDepartureTimingInformation(departure)
		if err != nil {
			result = "error"
			break
		}

		if info.MinutesUntilDeparture < 0 && i == len(m.Departures)-1 {
			result = "No data available"
			break
		}
		departureTimeString := info.DepartureTime.In(info.Location).Format("03:04 PM")
		isPollingStr := map[bool]string{true: " (polling)", false: ""}[m.isPolling]
		result = fmt.Sprintf(
			"Train in %.0fmin (%s)%s",
			info.MinutesUntilDeparture,
			departureTimeString,
			isPollingStr,
		)
	}
	return result
}
