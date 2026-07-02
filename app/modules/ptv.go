package modules

import (
	"fmt"
	"sync"
	"time"

	ptv "github.com/rolandwarburton/ptv-go/pkg"
	logger "github.com/rolandwarburton/sway-status-line/app/logger"
	types "github.com/rolandwarburton/sway-status-line/app/types"
)

type PublicTransport struct {
	Module
	mu            sync.RWMutex
	Departures    []ptv.Departure
	NextDeparture time.Time
	Polled        bool
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

	var nextDeparture time.Time
	if len(departures) > 0 {
		info, err := GetDepartureTimingInformation(departures[0])
		if err != nil {
			logger.Info(err.Error())
			return
		}
		nextDeparture = info.DepartureTime
	} else {
		logger.Info("no departures returned")
	}

	m.mu.Lock()
	m.Departures = departures
	m.NextDeparture = nextDeparture
	m.Polled = true
	m.mu.Unlock()
}

func (m *PublicTransport) Init(config types.ModulePtv) {
	logger.Info(
		fmt.Sprintf(
			"set train route to %s going towards %s",
			config.RouteName, config.DirectionName,
		),
	)
	// Only override the library's env-derived secrets (PTV_KEY/PTV_DEVID) when
	// the config actually supplies both, so a blank config doesn't clobber them.
	if config.PTVKEY != "" && config.PTVDEVID != "" {
		ptv.SetPTVSecrets(config.PTVKEY, config.PTVDEVID)
	}
	m.Enabled = true
	go func() {
		for {
			m.mu.RLock()
			next := m.NextDeparture
			m.mu.RUnlock()

			// poll if we have no departure yet or the cached one is in the past
			if next.IsZero() || time.Now().After(next) {
				m.poll(config)
			}

			time.Sleep(4 * time.Second)
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
	m.mu.RLock()
	polled := m.Polled
	next := m.NextDeparture
	haveDeparture := len(m.Departures) > 0
	var departure ptv.Departure
	if haveDeparture {
		departure = m.Departures[0]
	}
	m.mu.RUnlock()

	if !polled {
		return "Loading..."
	}

	if !haveDeparture {
		return "no departures"
	}

	if time.Now().After(next) {
		return "loading..."
	}

	info, err := GetDepartureTimingInformation(departure)
	if err != nil {
		return "error"
	}

	departureTimeString := info.DepartureTime.In(info.Location).Format("03:04 PM")
	return fmt.Sprintf(
		"Train in %.0fmin (%s)",
		info.MinutesUntilDeparture,
		departureTimeString,
	)
}
