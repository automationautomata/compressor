package utiles

import (
	"sync"
)

type SafeMap[T comparable, F any] struct {
	data map[T]F
	mu   sync.RWMutex
}

func newSafeMap[T comparable, F any](data map[T]F) *SafeMap[T, F] {
	return &SafeMap[T, F]{data: data}
}

func NewSafeMap[T comparable, F any]() *SafeMap[T, F] {
	return newSafeMap(make(map[T]F))
}

func (m *SafeMap[T, F]) Load(key T) (F, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *SafeMap[T, F]) Store(key T, value F) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SafeMap[T, F]) Range(f func(k T, v F) (pass bool)) {
	m.mu.Lock()
	data := make(map[T]F)
	for k, v := range m.data {
		data[k] = v
	}
	m.mu.Unlock()
	for k, v := range data {
		if !f(k, v) {
			return
		}
	}
}

func (m *SafeMap[T, F]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.data)
}

func (m *SafeMap[T, F]) ToMap() map[T]F {
	data := make(map[T]F)

	m.Range(func(k T, v F) bool {
		data[k] = v
		return true
	})
	return data
}

func FromMap[T comparable, F any](src map[T]F) *SafeMap[T, F] {
	data := make(map[T]F)
	for k, v := range src {
		data[k] = v
	}
	return newSafeMap(data)
}

func SafeMapCopy[T comparable, F any](sources ...*SafeMap[T, F]) *SafeMap[T, F] {
	data := make(map[T]F)

	for i := range sources {
		sources[i].Range(func(k T, v F) bool {
			data[k] = v
			return true
		})
	}
	return &SafeMap[T, F]{data: data}
}
