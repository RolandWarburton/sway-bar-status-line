package logger

import (
	journal "github.com/coreos/go-systemd/v22/journal"
)

func Info(message string) {
	journal.Send(message, journal.PriInfo, map[string]string{"SYSLOG_IDENTIFIER": "status_bar"})
}

func Alert(message string) {

	journal.Send(message, journal.PriAlert, map[string]string{"SYSLOG_IDENTIFIER": "status_bar"})
}
