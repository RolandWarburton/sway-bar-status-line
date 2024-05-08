package modules

import (
	"time"
)

type Time struct {
	Module
}

func (m *Time) Init() {
	m.Enabled = true
}

func (m *Time) Run() string {
	return time.Now().Format("2006-01-02 03:04:05 PM")
}
