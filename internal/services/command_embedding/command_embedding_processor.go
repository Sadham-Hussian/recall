package command_embedding

import (
	"fmt"
	"recall/internal/embedding"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
	"sort"
)

type CommandEmbeddingProcessor struct {
	CommandExecutionRepo      *repositories.CommandExecutionRepository
	CommandEmbeddingRepo      *repositories.CommandEmbeddingRepository
	CommandEmbeddingQueueRepo *repositories.CommandEmbeddingQueueRepository
	Embedder                  embedding.Embedder
	model                     string
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
		CommandExecutionRepo:      repositories.NewCommandExecutionRepository(db),
		CommandEmbeddingRepo:      repositories.NewCommandEmbeddingRepository(db),
		CommandEmbeddingQueueRepo: repositories.NewCommandEmbeddingQueueRepository(db),
		Embedder:                  embedder,
		model:                     model,
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

func (s *CommandEmbeddingProcessor) Search(query string, topK int) ([]models.SearchResult, error) {
	// 1. Embed query
	queryVec, err := s.Embedder.Embed(query)
	if err != nil {
		return nil, err
	}

	// 2. Fetch stored embeddings
	data, err := s.CommandEmbeddingRepo.FetchAllEmbeddings(s.model)
	if err != nil {
		return nil, err
	}

	var results []models.SearchResult

	// 3. Compute similarity
	for _, item := range data {
		vec, err := embedding.BytesToFloats(item.Vector)
		if err != nil {
			continue // skip corrupted rows
		}

		score := embedding.CosineSimilarity(queryVec, vec)

		results = append(results, models.SearchResult{
			Command: item.Command,
			Score:   score,
		})
	}

	// 4. Sort DESC
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 5. Top K
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}
