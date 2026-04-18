package repositories

import (
	"recall/internal/storage/models"
	"strings"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// Overview returns aggregate counts across all command executions.
// When sinceTs > 0 it scopes to commands recorded after that timestamp.
func (r *StatsRepository) Overview(sinceTs int64) (*models.OverviewStats, error) {
	var s models.OverviewStats

	query := `
		SELECT
			COUNT(*) as total_commands,
			COUNT(DISTINCT command) as unique_commands,
			COUNT(DISTINCT session_id) as total_sessions,
			COALESCE(SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END), 0) as success_count,
			COALESCE(MIN(timestamp), 0) as first_timestamp
		FROM command_executions
	`
	if sinceTs > 0 {
		query += " WHERE timestamp >= ?"
		err := r.db.Raw(query, sinceTs).Scan(&s).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := r.db.Raw(query).Scan(&s).Error
		if err != nil {
			return nil, err
		}
	}

	s.FailureCount = s.TotalCommands - s.SuccessCount
	return &s, nil
}

// TopCommands returns the N most frequently executed commands.
func (r *StatsRepository) TopCommands(sinceTs int64, limit int) ([]models.CommandCount, error) {
	var results []models.CommandCount
	q := r.db.Table("command_executions").
		Select("command, COUNT(*) as count").
		Where("command NOT LIKE 'recall%' AND command NOT LIKE './recall%'").
		Group("command").
		Order("count DESC").
		Limit(limit)
	if sinceTs > 0 {
		q = q.Where("timestamp >= ?", sinceTs)
	}
	err := q.Find(&results).Error
	return results, err
}

// TopCommandGroups groups commands by the first token (binary name) and
// returns the N most frequent groups with their top subcommands.
// e.g. "git" → (1204) — status, diff, stash, add, commit
func (r *StatsRepository) TopCommandGroups(sinceTs int64, limit int) ([]models.CommandGroup, error) {
	// Step 1: get top groups by first token
	groupQuery := `
		SELECT
			CASE
				WHEN INSTR(command, ' ') > 0 THEN SUBSTR(command, 1, INSTR(command, ' ') - 1)
				ELSE command
			END as "group",
			COUNT(*) as count
		FROM command_executions
		WHERE command NOT LIKE 'recall%' AND command NOT LIKE './recall%'
	`
	if sinceTs > 0 {
		groupQuery += " AND timestamp >= ?"
	}
	groupQuery += ` GROUP BY "group" ORDER BY count DESC LIMIT ?`

	var groups []models.CommandGroup
	var err error
	if sinceTs > 0 {
		err = r.db.Raw(groupQuery, sinceTs, limit).Scan(&groups).Error
	} else {
		err = r.db.Raw(groupQuery, limit).Scan(&groups).Error
	}
	if err != nil {
		return nil, err
	}

	// Step 2: for each group, find top 5 subcommands (second token)
	subQuery := `
		SELECT
			CASE
				WHEN INSTR(SUBSTR(command, LENGTH(?) + 2), ' ') > 0
				THEN SUBSTR(SUBSTR(command, LENGTH(?) + 2), 1, INSTR(SUBSTR(command, LENGTH(?) + 2), ' ') - 1)
				ELSE SUBSTR(command, LENGTH(?) + 2)
			END as sub,
			COUNT(*) as cnt
		FROM command_executions
		WHERE command LIKE ? || ' %'
		GROUP BY sub
		ORDER BY cnt DESC
		LIMIT 5
	`

	for i, g := range groups {
		var subs []struct {
			Sub string
			Cnt int64
		}
		r.db.Raw(subQuery, g.Group, g.Group, g.Group, g.Group, g.Group).Scan(&subs)
		parts := make([]string, 0, len(subs))
		for _, s := range subs {
			if s.Sub != "" {
				parts = append(parts, s.Sub)
			}
		}
		if len(parts) > 0 {
			groups[i].Subcommands = strings.Join(parts, ", ")
		}
	}

	return groups, nil
}

// MostFailed returns commands with the highest absolute failure count,
// filtered to commands that have been run at least minRuns times.
func (r *StatsRepository) MostFailed(sinceTs int64, limit, minRuns int) ([]models.FailedCommand, error) {
	var results []models.FailedCommand
	q := r.db.Table("command_executions").
		Select(`command,
			COUNT(*) as total_count,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) as failure_count`).
		Where("command NOT LIKE 'recall%' AND command NOT LIKE './recall%'").
		Group("command").
		Having("failure_count > 0 AND total_count >= ?", minRuns).
		Order("failure_count DESC").
		Limit(limit)
	if sinceTs > 0 {
		q = q.Where("timestamp >= ?", sinceTs)
	}
	err := q.Find(&results).Error
	return results, err
}

// TopDirectories returns the N most active working directories.
func (r *StatsRepository) TopDirectories(sinceTs int64, limit int) ([]models.DirectoryCount, error) {
	var results []models.DirectoryCount
	q := r.db.Table("command_executions").
		Select("cwd, COUNT(*) as count").
		Where("cwd != ''").
		Group("cwd").
		Order("count DESC").
		Limit(limit)
	if sinceTs > 0 {
		q = q.Where("timestamp >= ?", sinceTs)
	}
	err := q.Find(&results).Error
	return results, err
}

// ActivityByDay returns command counts grouped by day of week (0=Sun, 6=Sat).
// Always scoped to the last 7 days for a meaningful weekly view.
func (r *StatsRepository) ActivityByDay(sinceTs int64) ([]models.DayActivity, error) {
	var results []models.DayActivity
	q := r.db.Table("command_executions").
		Select("CAST(strftime('%w', timestamp, 'unixepoch', 'localtime') AS INTEGER) as day_of_week, COUNT(*) as count").
		Group("day_of_week").
		Order("day_of_week ASC")
	if sinceTs > 0 {
		q = q.Where("timestamp >= ?", sinceTs)
	}
	err := q.Find(&results).Error
	return results, err
}

// ActivityByHour returns the top N busiest hours of the day.
func (r *StatsRepository) ActivityByHour(sinceTs int64, limit int) ([]models.HourActivity, error) {
	var results []models.HourActivity
	q := r.db.Table("command_executions").
		Select("CAST(strftime('%H', timestamp, 'unixepoch', 'localtime') AS INTEGER) as hour, COUNT(*) as count").
		Group("hour").
		Order("count DESC").
		Limit(limit)
	if sinceTs > 0 {
		q = q.Where("timestamp >= ?", sinceTs)
	}
	err := q.Find(&results).Error
	return results, err
}
