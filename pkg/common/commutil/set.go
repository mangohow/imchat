package commutil

import "sync"

type ConcurrentSet[T comparable] struct {
	mux sync.RWMutex
	m map[T]struct{}
}

func NewConcurrentSet[T comparable]() *ConcurrentSet[T] {
	return &ConcurrentSet[T]{
		m: make(map[T]struct{}),
	}
}

func (s *ConcurrentSet[T]) Set(t T) {
	s.mux.Lock()
	s.m[t] = struct{}{}
	s.mux.Unlock()
}

func (s *ConcurrentSet[T]) IsSet(t T) (ok bool) {
	s.mux.RLock()
	_, ok = s.m[t]
	s.mux.RUnlock()
	return
}

func (s *ConcurrentSet[T]) Del(t T) {
	s.mux.Lock()
	delete(s.m, t)
	s.mux.Unlock()
}
