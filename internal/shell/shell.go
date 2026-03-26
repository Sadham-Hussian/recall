package shell

import "time"

type Entry struct {
	Command   string
	Timestamp time.Time
}

type Shell interface {
	Name() string
	HistoryFile() string
	ReadHistory() ([]Entry, error)
}
