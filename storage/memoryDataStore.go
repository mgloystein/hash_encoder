package storage

import (
	"fmt"
	"sync"
)

func newMemoryStorage() DataStore {
	var lock sync.RWMutex
	return &memoryDataStore{
		store: make(map[int]string),
		lock:  &lock,
		last:  0,
	}
}

type memoryDataStore struct {
	store map[int]string
	lock  *sync.RWMutex
	last  int
}

func (m *memoryDataStore) Get(itemID int) (string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	item, prs := m.store[itemID]
	if !prs {
		return item, fmt.Errorf("No item found at %d", itemID)
	}
	return item, nil
}

func (m *memoryDataStore) Reserve() Persistable {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.last = m.last + 1

	return &memoryPersistance{
		memoryDataStore: m,
		id:              m.last,
	}
}

type memoryPersistance struct {
	*memoryDataStore
	id        int
	persisted bool
}

func (p *memoryPersistance) ID() int {
	return p.id
}

func (p *memoryPersistance) Persist(item string) error {
	if p.persisted {
		return nil
	}
	_, prs := p.store[p.id]
	if prs {
		return fmt.Errorf("Item id %d was reserved", p.id)
	}

	p.store[p.id] = item
	p.persisted = true

	return nil
}
