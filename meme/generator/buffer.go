package generator

import (
	"bytes"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(buffer)
	},
}

type buffer struct {
	bytes.Buffer
}

func newBuffer() *buffer {
	return bufferPool.Get().(*buffer)
}

func (b *buffer) Recycle() {
	b.Reset()
	bufferPool.Put(b)
}
