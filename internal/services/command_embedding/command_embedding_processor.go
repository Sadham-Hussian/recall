package command_embedding

import (
	"fmt"
	"recall/internal/embedding"
	"recall/internal/search"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
	"sort"
)

type CommandEmbeddingProcessor struct {
	CommandExecutionRepo       *repositories.CommandExecutionRepository
	CommandExecutionSearchRepo *repositories.CommandExecutionSearchRepository
	CommandEmbeddingRepo       *repositories.CommandEmbeddingRepository
	CommandEmbeddingQueueRepo  *repositories.CommandEmbeddingQueueRepository
	Embedder                   embedding.Embedder
	model                      string
}

func NewEmbeddingProcessor(
	embedder embedding.Embedder,
	model string,
) (*CommandEmbeddingProcessor, error) {

	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &CommandEmbeddingProcessor{
		CommandExecutionRepo:       repositories.NewCommandExecutionRepository(db),
		CommandExecutionSearchRepo: repositories.NewCommandExecutionSearchRepository(db),
		CommandEmbeddingRepo:       repositories.NewCommandEmbeddingRepository(db),
		CommandEmbeddingQueueRepo:  repositories.NewCommandEmbeddingQueueRepository(db),
		Embedder:                   embedder,
		model:                      model,
	}, nil
}

func (p *CommandEmbeddingProcessor) Process(batchSize int) error {
	items, err := p.CommandExecutionRepo.FetchForEmbedding(batchSize)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No pending embeddings.")
		return nil
	}

	fmt.Printf("Processing %d commands...\n", len(items))

	var processedIDs []int64

	for _, item := range items {
		vec, err := p.Embedder.Embed(item.Command)
		if err != nil {
			fmt.Printf("Embedding failed for ID %d: %v\n", item.ID, err)
			continue // skip but don't delete from queue
		}

		bytes, err := embedding.FloatsToBytes(vec)
		if err != nil {
			fmt.Printf("Failed to convert floats to bytes for ID %d: %v\n", item.ID, err)
			continue
		}

		err = p.CommandEmbeddingRepo.InsertEmbedding(models.CommandEmbedding{
			CommandExecutionID: item.ID,
			Model:              p.model,
			Dimensions:         len(vec),
			Embedding:          bytes,
		})

		if err != nil {
			fmt.Printf("Insert failed for ID %d: %v\n", item.ID, err)
			continue
		}

		processedIDs = append(processedIDs, int64(item.ID))
	}

	err = p.CommandEmbeddingQueueRepo.DeleteFromQueue(processedIDs)
	if err != nil {
		return err
	}

	fmt.Printf("Processed: %d\n", len(processedIDs))
	fmt.Printf("Remaining in queue: %d\n", len(items)-len(processedIDs))

	return nil
}

func (p *CommandEmbeddingProcessor) HasPendingEmbeddings() (bool, error) {
	count, err := p.CommandEmbeddingQueueRepo.CountQueue()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *CommandEmbeddingProcessor) Search(query string, topK int, ftsSearchLimit, embeddingCandidateLimit int) ([]models.SearchResult, error) {
	// 1. Embed query
	queryVec, err := s.Embedder.Embed(query)
	if err != nil {
		return nil, err
	}

	// 2. Fetch semantic candidates
	data, err := s.CommandEmbeddingRepo.FetchAllEmbeddings(s.model, embeddingCandidateLimit)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]*models.SearchResult)

	// 3. Semantic scoring
	for _, item := range data {
		vec, err := embedding.BytesToFloats(item.Vector)
		if err != nil {
			continue
		}

		score := embedding.CosineSimilarity(queryVec, vec)

		resultMap[item.Command] = &models.SearchResult{
			Command:       item.Command,
			SemanticScore: score,
		}
	}

	// 4. FTS search
	ftsQuery := search.BuildFTSQueryForSemanticSearch(query)
	if ftsQuery != "" {
		ftsResults, err := s.CommandExecutionSearchRepo.FTSSearch(ftsQuery, ftsSearchLimit)
		if err == nil {

			// 1. Deduplicate + keep BEST rank (lower bm25 = better)
			bestFTS := make(map[string]float64)

			for _, fts := range ftsResults {
				existingRank, ok := bestFTS[fts.Command]
				if !ok || fts.Rank < existingRank {
					bestFTS[fts.Command] = fts.Rank
				}
			}

			// 2. Merge into resultMap
			for cmd, rank := range bestFTS {
				r, exists := resultMap[cmd]
				if !exists {
					r = &models.SearchResult{
						Command: cmd,
					}
					resultMap[cmd] = r
				}

				score := normalizeFTSScore(rank)

				if score > r.FTSScore {
					r.FTSScore = score
				}
			}
		}
	}

	// 5. Final scoring (initial version)
	var results []models.SearchResult

	for _, r := range resultMap {
		r.FinalScore =
			0.7*r.SemanticScore +
				0.3*r.FTSScore

		results = append(results, *r)
	}

	// 6. Sort
	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})

	// 7. Top K
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func normalizeFTSScore(rank float64) float64 {
	if rank <= 0 {
		return 0
	}
	return 1 / (1 + rank) // simple inversion
}
