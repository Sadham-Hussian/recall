package stats

import (
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
)

type StatsService struct {
	StatsRepo *repositories.StatsRepository
}

func NewStatsService() (*StatsService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &StatsService{
		StatsRepo: repositories.NewStatsRepository(db),
	}, nil
}

func (s *StatsService) Overview(sinceTs int64) (*models.OverviewStats, error) {
	return s.StatsRepo.Overview(sinceTs)
}

func (s *StatsService) TopCommands(sinceTs int64, limit int) ([]models.CommandCount, error) {
	return s.StatsRepo.TopCommands(sinceTs, limit)
}

func (s *StatsService) TopCommandGroups(sinceTs int64, limit int) ([]models.CommandGroup, error) {
	return s.StatsRepo.TopCommandGroups(sinceTs, limit)
}

func (s *StatsService) MostFailed(sinceTs int64, limit, minRuns int) ([]models.FailedCommand, error) {
	return s.StatsRepo.MostFailed(sinceTs, limit, minRuns)
}

func (s *StatsService) TopDirectories(sinceTs int64, limit int) ([]models.DirectoryCount, error) {
	return s.StatsRepo.TopDirectories(sinceTs, limit)
}

func (s *StatsService) ActivityByDay(sinceTs int64) ([]models.DayActivity, error) {
	return s.StatsRepo.ActivityByDay(sinceTs)
}

func (s *StatsService) ActivityByHour(sinceTs int64, limit int) ([]models.HourActivity, error) {
	return s.StatsRepo.ActivityByHour(sinceTs, limit)
}
