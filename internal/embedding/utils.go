package embedding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func BytesToFloats(b []byte) ([]float32, error) {
	if len(b)%4 != 0 {
		return nil, fmt.Errorf("invalid byte slice length: %d", len(b))
	}

	n := len(b) / 4
	floats := make([]float32, n)

	for i := 0; i < n; i++ {
		bits := binary.LittleEndian.Uint32(b[i*4:])
		floats[i] = math.Float32frombits(bits)
	}

	return floats, nil
}

func FloatsToBytes(floats []float32) ([]byte, error) {
	buf := new(bytes.Buffer)

	for _, f := range floats {
		if err := binary.Write(buf, binary.LittleEndian, f); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
