package models

type OverviewStats struct {
	TotalCommands  int64
	UniqueCommands int64
	TotalSessions  int64
	SuccessCount   int64
	FailureCount   int64
	FirstTimestamp int64
}

type CommandCount struct {
	Command string
	Count   int64
}

type CommandGroup struct {
	Group       string
	Count       int64
	Subcommands string
}

type FailedCommand struct {
	Command      string
	TotalCount   int64
	FailureCount int64
}

type DirectoryCount struct {
	CWD   string
	Count int64
}

type DayActivity struct {
	DayOfWeek int
	Count     int64
}

type HourActivity struct {
	Hour  int
	Count int64
}
