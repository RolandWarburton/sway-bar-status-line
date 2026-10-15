package modules

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/pilebones/go-udev/netlink"
)

const (
	logindDest    = "org.freedesktop.login1"
	logindPath    = "/org/freedesktop/login1"
	logindManager = "org.freedesktop.login1.Manager"
)

// lid switch actions that leave the machine running
var lidAwakeActions = []string{"ignore", "lock"}

type Lid struct {
	Module
	mu  sync.RWMutex
	obj dbus.BusObject

	// HandleLidSwitch* are const in logind, so they are only read once
	action, dockedAction, externalAction string

	inhibited  bool // BlockInhibited contains handle-lid-switch
	docked     bool
	onExternal bool

	err error // sticky: once set the module stays in LID ERR
}

// Init reads the current state, then watches for changes instead of polling:
//   - BlockInhibited: logind PropertiesChanged signal
//   - Docked / OnExternalPower: logind sends no signal for these, so they are
//     re-read on drm / power_supply kernel uevents. The kernel stores the new
//     connector / supply state before sending the uevent, so no settle delay
//     is needed.
//
// There is no fallback poll. Any failure is shown as LID ERR rather than
// silently going stale.
func (m *Lid) Init() {
	m.Enabled = true

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		m.fail(err)
		return
	}
	m.obj = conn.Object(logindDest, logindPath)

	// subscribe before the initial read so no change slips between them
	signals := make(chan *dbus.Signal, 8)
	conn.Signal(signals)
	if err := conn.AddMatchSignal(
		dbus.WithMatchSender(logindDest),
		dbus.WithMatchObjectPath(logindPath),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		m.fail(err)
		return
	}

	uevents := &netlink.UEventConn{}
	if err := uevents.Connect(netlink.KernelEvent); err != nil {
		m.fail(err)
		return
	}
	queue := make(chan netlink.UEvent)
	errs := make(chan error)
	go uevents.MonitorWithContext(context.Background(), queue, errs, &netlink.RuleDefinitions{
		Rules: []netlink.RuleDefinition{{Env: map[string]string{"SUBSYSTEM": "^(drm|power_supply)$"}}},
	})

	if err := m.readAll(); err != nil {
		m.fail(err)
		return
	}

	go func() { m.fail(m.watch(signals, queue, errs)) }()
}

// Run reports one of three states:
//
//	┌─────────────────────────────────────┬─────────┐
//	│              Condition              │ Output  │
//	├─────────────────────────────────────┼─────────┤
//	│ Bus, uevent monitor or read failed  │ LID ERR │
//	├─────────────────────────────────────┼─────────┤
//	│ Lid close would sleep               │ ⏾       │
//	├─────────────────────────────────────┼─────────┤
//	│ Inhibited, or action is ignore/lock │ 󰅶       │
//	└─────────────────────────────────────┴─────────┘
func (m *Lid) Run() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return "LID ERR"
	}
	if m.sleeps() {
		return "⏾"
	}
	return "󰅶"
}

func (m *Lid) fail(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err == nil {
		m.err = err
	}
}

// sleeps mirrors logind's decision for a lid close: a block inhibitor on
// handle-lid-switch (e.g. lid-inhibit) wins, otherwise the configured action
// for the current docked / external power state applies. Caller holds mu.
func (m *Lid) sleeps() bool {
	if m.inhibited {
		return false
	}
	action := m.action
	// an empty ExternalPower action means fall back to HandleLidSwitch
	if m.externalAction != "" && m.onExternal {
		action = m.externalAction
	}
	// docked takes precedence over external power in logind
	if m.docked {
		action = m.dockedAction
	}
	return !slices.Contains(lidAwakeActions, action)
}

func (m *Lid) watch(signals chan *dbus.Signal, queue chan netlink.UEvent, errs chan error) error {
	for {
		select {
		case sig, ok := <-signals:
			// godbus closes the channel when the bus connection drops
			if !ok {
				return errors.New("system bus connection closed")
			}
			// the bus also delivers its own NameAcquired signal here
			if sig.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
				continue
			}
			changed, _ := sig.Body[1].(map[string]dbus.Variant)
			v, ok := changed["BlockInhibited"]
			if !ok {
				continue
			}
			var blocked string
			if err := v.Store(&blocked); err != nil {
				return err
			}
			m.mu.Lock()
			m.inhibited = lidSwitchBlocked(blocked)
			m.mu.Unlock()
		case _, ok := <-queue:
			if !ok {
				return errors.New("uevent monitor stopped")
			}
			if err := m.readPower(); err != nil {
				return err
			}
		case err := <-errs:
			return err
		}
	}
}

func lidSwitchBlocked(blocked string) bool {
	return slices.Contains(strings.Split(blocked, ":"), "handle-lid-switch")
}

func (m *Lid) readAll() error {
	var action, dockedAction, externalAction, blocked string
	for name, dst := range map[string]*string{
		"HandleLidSwitch":              &action,
		"HandleLidSwitchDocked":        &dockedAction,
		"HandleLidSwitchExternalPower": &externalAction,
		"BlockInhibited":               &blocked,
	} {
		if err := getProperty(m.obj, name, dst); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.action, m.dockedAction, m.externalAction = action, dockedAction, externalAction
	m.inhibited = lidSwitchBlocked(blocked)
	m.mu.Unlock()

	return m.readPower()
}

func (m *Lid) readPower() error {
	var docked, onExternal bool
	if err := getProperty(m.obj, "Docked", &docked); err != nil {
		return err
	}
	if err := getProperty(m.obj, "OnExternalPower", &onExternal); err != nil {
		return err
	}

	m.mu.Lock()
	m.docked, m.onExternal = docked, onExternal
	m.mu.Unlock()
	return nil
}

func getProperty(obj dbus.BusObject, name string, dst any) error {
	v, err := obj.GetProperty(logindManager + "." + name)
	if err != nil {
		return err
	}
	return v.Store(dst)
}
