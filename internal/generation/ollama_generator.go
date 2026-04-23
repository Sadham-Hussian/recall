package generation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaGenerator struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaGenerator(baseURL, model string, timeoutSeconds int) *OllamaGenerator {
	return &OllamaGenerator{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
}

// Generate sends a prompt to Ollama's /api/generate endpoint with
// streaming enabled. Returns a ReadCloser that yields the response
// text as it arrives. Each read returns decoded text chunks (not raw
// JSON) — the caller can io.Copy straight to stdout.
func (g *OllamaGenerator) Generate(ctx context.Context, systemPrompt, userPrompt string) (io.ReadCloser, error) {
	reqBody := ollamaGenerateRequest{
		Model:  g.model,
		Prompt: userPrompt,
		System: systemPrompt,
		Stream: true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama returned %s", resp.Status)
	}

	return &streamReader{body: resp.Body, decoder: json.NewDecoder(resp.Body)}, nil
}

// streamReader decodes Ollama's newline-delimited JSON stream and
// yields just the text content. Each call to Read returns decoded
// response tokens — the caller sees plain text, not JSON.
type streamReader struct {
	body    io.ReadCloser
	decoder *json.Decoder
	buf     []byte
}

type ollamaStreamChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (r *streamReader) Read(p []byte) (int, error) {
	// Drain any leftover buffer from previous decode
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}

	// Decode next JSON chunk from the stream
	var chunk ollamaStreamChunk
	if err := r.decoder.Decode(&chunk); err != nil {
		return 0, err
	}
	if chunk.Done {
		return 0, io.EOF
	}

	// Copy decoded text into caller's buffer
	data := []byte(chunk.Response)
	n := copy(p, data)
	if n < len(data) {
		r.buf = data[n:]
	}
	return n, nil
}

func (r *streamReader) Close() error {
	return r.body.Close()
}
