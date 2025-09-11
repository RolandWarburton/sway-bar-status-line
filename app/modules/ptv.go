package modules

import (
	"fmt"
	"time"

	ptv "github.com/rolandwarburton/ptv-go/pkg"
	logger "github.com/rolandwarburton/sway-status-line/app/logger"
	types "github.com/rolandwarburton/sway-status-line/app/types"
)

type PublicTransport struct {
	Module
	Departures    []ptv.Departure
	NextDeparture time.Time
}

func (m *PublicTransport) poll(config types.ModulePtv) {
	// avoid sending more than one request at a time
	logger.Info("polling PTV")
	departures, err := ptv.DeparturesAction(config.RouteName, config.StopName, config.DirectionName, 1, "Australia/Sydney")
	if err != nil {
		logger.Alert(fmt.Sprintf("error polling: %s", err.Error()))
		logger.Info(err.Error())
		return
	}
	m.Departures = departures
	nextDeparture, err := GetDepartureTimingInformation(departures[0])
	if err != nil {
		logger.Info(err.Error())
		return
	}
	m.NextDeparture = nextDeparture.DepartureTime
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
	go func() {
		for {
			if m.NextDeparture.IsZero() {
				m.poll(config)
			}
			// if the next departure is in the past
			if time.Now().Compare(m.NextDeparture) == 1 {
				m.poll(config)
			}

			// sleep 5min
			sleepTime := 4 * time.Second
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
		return "Loading..."
	}

	if time.Now().Compare(m.NextDeparture) == 1 {
		return "loading..."
	}

	var result string
	departure := m.Departures[0]
	info, err := GetDepartureTimingInformation(departure)
	if err != nil {
		result = "error"
	}

	departureTimeString := info.DepartureTime.In(info.Location).Format("03:04 PM")
	result = fmt.Sprintf(
		"Train in %.0fmin (%s)",
		info.MinutesUntilDeparture,
		departureTimeString,
	)

	return result
}
