package cache

import (
	"context"
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

type MemoryService struct {
	items map[string]item
	mu    sync.RWMutex
}

func NewMemoryService() *MemoryService {
	return &MemoryService{
		items: make(map[string]item),
	}
}

func (s *MemoryService) Get(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, found := s.items[key]
	if !found {
		return "", nil // Cache miss, return empty string as per Redis behavior often expected
	}

	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return "", nil
	}

	if str, ok := item.value.(string); ok {
		return str, nil
	}
	return "", nil
}

func (s *MemoryService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	s.items[key] = item{
		value:      value,
		expiration: exp,
	}
	return nil
}

func (s *MemoryService) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *MemoryService) FlushAll(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]item)
	return nil
}
