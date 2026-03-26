package models

type CommandExecution struct {
	ID        uint   `gorm:"primaryKey"`
	Command   string `gorm:"not null"`
	Timestamp int64  `gorm:"not null"`
	CWD       string `gorm:"column:cwd"`
	ExitCode  int    `gorm:"column:exit_code"`
	ShellPID  int    `gorm:"column:shell_pid"`
	SessionID string `gorm:"column:session_id"`
}

type Session struct {
	SessionID string
	Commands  []CommandExecution
}
