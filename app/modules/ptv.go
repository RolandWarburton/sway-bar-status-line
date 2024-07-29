package modules

import (
	"fmt"
	"math"
	"time"

	ptv "github.com/rolandwarburton/ptv-go/pkg"
	logger "github.com/rolandwarburton/sway-status-line/app/logger"
	types "github.com/rolandwarburton/sway-status-line/app/types"
)

type PublicTransport struct {
	Module
	Departures    []ptv.Departure
	NextDeparture time.Time
	LastPoll      time.Time
	nextPoll      time.Time
	isPolling     bool
}

func (m *PublicTransport) poll(config types.ModulePtv) {
	// avoid sending more than one request at a time
	defer func() {
		m.isPolling = false
		logger.Info(fmt.Sprintf("finished polling PTV (%d results)", len(m.Departures)))
	}()
	if m.isPolling {
		return
	}
	m.isPolling = true
	logger.Info("polling PTV")
	departures, err := ptv.DeparturesAction(config.RouteName, config.StopName, config.DirectionName, 3, "Australia/Sydney")
	if err != nil {
		logger.Alert(fmt.Sprintf("error polling: %s", err.Error()))
		return
	}
	m.Departures = departures
	m.LastPoll = time.Now()
	m.LastPoll = time.Now()
	m.isPolling = false
}

func (m *PublicTransport) Init(config types.ModulePtv) {
	logger.Info(
		fmt.Sprintf(
			"set train route to %s going towards %s",
			config.RouteName, config.DirectionName,
		),
	)
	ptv.SetPTVSecrets(config.PTVKEY, config.PTVDEVID)
	m.Enabled = true
	m.nextPoll = time.Now()
	go func() {
		for {
			m.poll(config)
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
	i := 0
	for result == "" && i <= len(m.Departures) {
		departure := m.Departures[i]
		info, err := GetDepartureTimingInformation(departure)
		if err != nil {
			result = "error"
			break
		}

		if info.MinutesUntilDeparture < 1 && i == len(m.Departures)-1 {
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
		i++
	}

	return result
}
