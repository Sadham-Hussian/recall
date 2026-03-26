package embedding

type Embedder interface {
	Embed(text string) ([]float32, error)
}
