package models

type SessionName struct {
	SessionID string `gorm:"primaryKey;column:session_id"`
	Name      string `gorm:"not null"`
}
