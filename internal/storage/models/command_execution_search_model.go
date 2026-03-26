package models

type HybridSearchResult struct {
	Command       string
	Count         int64
	LastTimestamp int64
	SuccessCount  int64
	Cwd           string
	SessionID     string
	Score         float64 `gorm:"-"`
	FuzzyScore    int     `gorm:"-"`
}
