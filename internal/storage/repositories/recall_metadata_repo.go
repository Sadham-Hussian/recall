package repositories

import (
	"strconv"

	"gorm.io/gorm"
)

type RecallMetadataRepository struct {
	db *gorm.DB
}

func NewRecallMetadataRepository(db *gorm.DB) *RecallMetadataRepository {
	return &RecallMetadataRepository{db: db}
}

func (r *RecallMetadataRepository) GetLastEmbeddedID() (int64, error) {
	var value string

	err := r.db.Raw(`
		SELECT value
		FROM recall_metadata
		WHERE key = 'last_embedded_command_id'
	`).Scan(&value).Error

	if err != nil {
		return 0, err
	}

	if value == "" {
		return 0, nil
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *RecallMetadataRepository) SetLastEmbeddedID(id int64) error {
	return r.db.Exec(`
		INSERT INTO recall_metadata (key, value)
		VALUES ('last_embedded_command_id', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, strconv.FormatInt(id, 10)).Error
}
