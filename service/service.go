package service

import (
	"context"
	"errors"

	"github.com/alekshiga/ozon-url-shortener/storage"
)

const maxGenerateAttempts = 5

type Encoder interface {
	Encode() (string, error)
}

type Service struct {
	storage storage.Storage
	encoder Encoder
}

func New(storage storage.Storage, encoder Encoder) *Service {
	return &Service{
		storage: storage,
		encoder: encoder,
	}
}

func (s *Service) Shorten(ctx context.Context, originalURL string) (string, error) {
	existingCode, err := s.storage.GetByOriginal(ctx, originalURL)

	if err == nil {
		return existingCode, nil
	}

	if !errors.Is(err, storage.ErrNotFound) {
		return "", err
	}

	for range maxGenerateAttempts {
		shortCode, err := s.encoder.Encode()
		if err != nil {
			return "", err
		}

		err = s.storage.Save(ctx, originalURL, shortCode)

		if err == nil {
			return shortCode, nil
		}

		if errors.Is(err, storage.ErrURLExists) {
			return s.storage.GetByOriginal(ctx, originalURL)
		}

		if errors.Is(err, storage.ErrCodeExists) {
			continue
		}

		return "", err
	}

	return "", errors.New("failed to generate unique short code")
}

func (s *Service) GetOriginal(ctx context.Context, shortCode string) (string, error) {
	return s.storage.GetByShort(ctx, shortCode)
}
