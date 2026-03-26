package command_execution

import (
	"fmt"
	"os"
	"recall/internal/config"
	"recall/internal/search"
	"recall/internal/storage"
	"sort"
	"strings"

	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
)

type CommandExecutionSearchService struct {
	CommandExecutionSearchRepository *repositories.CommandExecutionSearchRepository
	CommandExecutionRepository       *repositories.CommandExecutionRepository
}

func NewCommandExecutionSearchService() (*CommandExecutionSearchService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &CommandExecutionSearchService{
		CommandExecutionSearchRepository: repositories.NewCommandExecutionSearchRepository(db),
		CommandExecutionRepository:       repositories.NewCommandExecutionRepository(db),
	}, nil
}

func (s *CommandExecutionSearchService) Search(cfg *config.Config, args []string, searchLimit int) ([]models.HybridSearchResult, error) {
	rawQuery := strings.Join(args, " ")
	ftsQuery := search.BuildFTSQuery(rawQuery)

	if searchLimit == 0 {
		searchLimit = cfg.Search.DefaultSearchLimit
	}

	results, err := s.CommandExecutionSearchRepository.HybridSearch(ftsQuery, searchLimit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	if len(results) == 0 {

		results, err = s.CommandExecutionSearchRepository.FuzzySearch(rawQuery, cfg.Search.FuzzySearchLimit)
		if err != nil {
			return nil, fmt.Errorf("fuzzy search failed: %v", err)
		}

		results = search.FuzzySearchFilter(results, rawQuery)

		if len(results) == 0 {
			fmt.Println("No matching commands found.")
			return nil, nil
		}
	}

	currentDir, _ := os.Getwd()
	fuzzyQuery := strings.ToLower(strings.Join(args, " "))
	projectRoot := search.GetProjectRoot(currentDir)
	currentSessionID, _ := s.CommandExecutionRepository.GetCurrentSessionID()
	for i := range results {
		results[i].FuzzyScore = search.FindFuzzyScore(&results[i], fuzzyQuery)
		results[i].Score = search.ComputeHybridScore(&results[i], currentDir, projectRoot, currentSessionID)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Trim to requested limit
	if len(results) > searchLimit {
		results = results[:searchLimit]
	}

	return results, nil
}
