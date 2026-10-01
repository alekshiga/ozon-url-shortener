package memory

import (
	"context"
	"sync"

	"github.com/alekshiga/ozon-url-shortener/storage"
)

type Storage struct {
	mu sync.RWMutex

	byShort    map[string]string
	byOriginal map[string]string
}

func New() *Storage {
	return &Storage{
		byShort:    make(map[string]string),
		byOriginal: make(map[string]string),
	}
}

func (s *Storage) Save(_ context.Context, originalURL string, shortCode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byOriginal[originalURL]; exists {
		return storage.ErrURLExists
	}

	if _, exists := s.byShort[shortCode]; exists {
		return storage.ErrCodeExists
	}

	s.byOriginal[originalURL] = shortCode
	s.byShort[shortCode] = originalURL

	return nil
}

func (s *Storage) GetByShort(_ context.Context, shortCode string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, exists := s.byShort[shortCode]
	if !exists {
		return "", storage.ErrNotFound
	}

	return originalURL, nil
}

func (s *Storage) GetByOriginal(_ context.Context, originalURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortCode, exists := s.byOriginal[originalURL]
	if !exists {
		return "", storage.ErrNotFound
	}

	return shortCode, nil
}
