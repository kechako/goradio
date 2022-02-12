package audio

import "sync"

type Frame[T SampleType] struct {
	data []T
	pool *framePool[T]
}

func newFrame[T SampleType](size int, pool *framePool[T]) *Frame[T] {
	return &Frame[T]{
		data: make([]T, size),
		pool: pool,
	}
}

func (f *Frame[T]) Data() []T {
	return f.data
}

func (f *Frame[T]) Release() {
}

type framePool[T SampleType] struct {
	pool sync.Pool
}

func newFramePool[T SampleType](size int) *framePool[T] {
	p := &framePool[T]{}
	p.pool = sync.Pool{
		New: func() any { return newFrame[T](size, p) },
	}
	return p
}

func (p *framePool[T]) Get() *Frame[T] {
	return p.pool.Get().(*Frame[T])
}

func (p *framePool[T]) Put(frame *Frame[T]) {
	p.pool.Put(frame)
}
