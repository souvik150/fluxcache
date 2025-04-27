package cache

import (
	"container/list"
	"sync"
)

type entry struct {
	key   string
	value any
}

type MemoryCache struct {
	mu       sync.RWMutex 
	store    map[string]*list.Element
	eviction *list.List
	capacity int
}

func NewMemoryCache(cap int) *MemoryCache {
	return &MemoryCache{
		store:    make(map[string]*list.Element),
		eviction: list.New(),
		capacity: cap,
	}
}

func (m *MemoryCache) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if el, ok := m.store[key]; ok {
		m.eviction.MoveToFront(el)
		el.Value.(*entry).value = value
		return
	}

	if m.eviction.Len() >= m.capacity {
		m.removeOldest()
	}

	ent := &entry{key, value}
	el := m.eviction.PushFront(ent)
	m.store[key] = el
}

func (m *MemoryCache) Get(key string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if el, ok := m.store[key]; ok {
		m.eviction.MoveToFront(el)
		return el.Value.(*entry).value, true
	}
	return nil, false
}

func (m *MemoryCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if el, ok := m.store[key]; ok {
		m.removeElement(el)
	}
}

func (m *MemoryCache) removeOldest() {
	el := m.eviction.Back()
	if el != nil {
		m.removeElement(el)
	}
}

func (m *MemoryCache) removeElement(el *list.Element) {
	ent := el.Value.(*entry)
	delete(m.store, ent.key)
	m.eviction.Remove(el)
}
