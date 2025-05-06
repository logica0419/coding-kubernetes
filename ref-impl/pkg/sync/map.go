package sync

import "sync"

type Map[K comparable, V any] struct {
	internal sync.Map
}

func (m *Map[K, V]) Clear() {
	m.internal.Clear()
}

func (m *Map[K, V]) Delete(key K) {
	m.internal.Delete(key)
}

func (m *Map[K, V]) Load(key K) (V, bool) {
	v, ok := m.internal.Load(key)
	if v == nil || !ok {
		return *new(V), false
	}

	value, ok := v.(V)

	return value, ok
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.internal.Range(func(k, v any) bool {
		if k, ok := k.(K); ok {
			if v, ok := v.(V); ok {
				return f(k, v)
			}
		}

		return false
	})
}

func (m *Map[K, V]) Store(key K, value V) {
	m.internal.Store(key, value)
}
