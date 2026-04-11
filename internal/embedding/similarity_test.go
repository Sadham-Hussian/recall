package embedding

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	const epsilon = 1e-6

	tests := []struct {
		name string
		a    []float32
		b    []float32
		want float64
	}{
		{
			name: "identical vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{1, 0, 0},
			want: 1.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float32{1, 0},
			b:    []float32{0, 1},
			want: 0.0,
		},
		{
			name: "opposite vectors",
			a:    []float32{1, 0},
			b:    []float32{-1, 0},
			want: -1.0,
		},
		{
			name: "zero vector a",
			a:    []float32{0, 0},
			b:    []float32{1, 0},
			want: 0.0,
		},
		{
			name: "zero vector b",
			a:    []float32{1, 0},
			b:    []float32{0, 0},
			want: 0.0,
		},
		{
			name: "length mismatch",
			a:    []float32{1, 2, 3},
			b:    []float32{1, 2},
			want: 0.0,
		},
		{
			name: "non-unit identical vectors",
			a:    []float32{3, 4},
			b:    []float32{3, 4},
			want: 1.0,
		},
		{
			name: "non-unit perpendicular vectors",
			a:    []float32{0, 5},
			b:    []float32{5, 0},
			want: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CosineSimilarity(tc.a, tc.b)
			if math.Abs(got-tc.want) > epsilon {
				t.Errorf("CosineSimilarity(%v, %v) = %f, want %f", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
