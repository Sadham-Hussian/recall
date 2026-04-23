package explain

import (
	"context"
	"io"

	"recall/internal/generation"
)

const systemPrompt = `You are a terminal command expert. When given a shell command, explain what it does in plain English.

Rules:
- If the command has pipes or multiple parts, explain each part in numbered steps
- End with a one-line "Overall:" summary
- Be concise — no preamble, no "Sure!", no markdown formatting
- If the command is trivial (ls, cd, pwd), keep the explanation to one line`

type ExplainService struct {
	Generator generation.Generator
}

func NewExplainService(generator generation.Generator) *ExplainService {
	return &ExplainService{Generator: generator}
}

// Explain sends the command to the LLM and returns a streaming reader.
// The caller should io.Copy the reader to stdout for real-time output.
func (s *ExplainService) Explain(ctx context.Context, command string) (io.ReadCloser, error) {
	userPrompt := "Explain this shell command:\n\n" + command
	return s.Generator.Generate(ctx, systemPrompt, userPrompt)
}
