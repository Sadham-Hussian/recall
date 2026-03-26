package models

type RecallMetadata struct {
	Key   string `gorm:"column:key;primaryKey"`
	Value string `gorm:"column:value"`
}
