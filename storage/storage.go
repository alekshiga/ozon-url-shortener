package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound  = errors.New("url not found")
	ErrURLExists = errors.New("url already exists")
	// ErrCodeExists для защиты от коллизий
	ErrCodeExists = errors.New("short code already exists")
)

type Storage interface {
	Save(ctx context.Context, originalURL, shortCode string) error
	GetByShort(ctx context.Context, shortCode string) (string, error)
	GetByOriginal(ctx context.Context, originalURL string) (string, error)
}
