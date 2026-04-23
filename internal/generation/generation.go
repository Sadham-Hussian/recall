package generation

import (
	"context"
	"io"
)

// Generator sends a prompt to an LLM and returns a streaming response.
// The caller must close the returned ReadCloser when done.
type Generator interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (io.ReadCloser, error)
}
