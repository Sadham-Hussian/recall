package repositories

import (
	"recall/internal/storage/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SessionNameRepository struct {
	db *gorm.DB
}

func NewSessionNameRepository(db *gorm.DB) *SessionNameRepository {
	return &SessionNameRepository{db: db}
}

// SetName creates or updates the name for a session.
func (r *SessionNameRepository) SetName(sessionID, name string) error {
	sn := models.SessionName{SessionID: sessionID, Name: name}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&sn).Error
}

// GetName returns the name for a session, or empty string if unnamed.
func (r *SessionNameRepository) GetName(sessionID string) (string, error) {
	var sn models.SessionName
	err := r.db.Where("session_id = ?", sessionID).First(&sn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return sn.Name, nil
}

// GetNames returns names for a list of session IDs as a map.
// Useful for batch lookups when displaying multiple sessions.
func (r *SessionNameRepository) GetNames(sessionIDs []string) (map[string]string, error) {
	var names []models.SessionName
	err := r.db.Where("session_id IN ?", sessionIDs).Find(&names).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(names))
	for _, n := range names {
		m[n.SessionID] = n.Name
	}
	return m, nil
}

// GetSessionIDByName looks up a session ID by its name.
func (r *SessionNameRepository) GetSessionIDByName(name string) (string, error) {
	var sn models.SessionName
	err := r.db.Where("name = ?", name).First(&sn).Error
	if err != nil {
		return "", err
	}
	return sn.SessionID, nil
}
