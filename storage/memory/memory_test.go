package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/alekshiga/ozon-url-shortener/storage"
)

func TestStorage_SaveAndGet(t *testing.T) {
	s := New()
	ctx := context.Background()

	originalURL := "https://google.com"
	shortCode := "1234512345"

	err := s.Save(ctx, originalURL, shortCode)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	gotURL, err := s.GetByShort(ctx, shortCode)
	if err != nil {
		t.Fatalf("GetByShort() returned error: %v", err)
	}

	if gotURL != originalURL {
		t.Errorf("expected %q, got %q", originalURL, gotURL)
	}

	gotCode, err := s.GetByOriginal(ctx, originalURL)
	if err != nil {
		t.Fatalf("GetByOriginal() returned error: %v", err)
	}

	if gotCode != shortCode {
		t.Errorf("expected %q, got %q", shortCode, gotCode)
	}
}

func TestStorage_URLAlreadyExists(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.Save(ctx, "https://google.com", "AAAAAAAAAA")
	if err != nil {
		t.Fatalf("first Save() returned error: %v", err)
	}

	err = s.Save(ctx, "https://google.com", "BBBBBBBBBB")

	if !errors.Is(err, storage.ErrURLExists) {
		t.Errorf("expected ErrURLExists, got %v", err)
	}
}

func TestStorage_CodeAlreadyExists(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.Save(ctx, "https://google.com", "AAAAAAAAAA")
	if err != nil {
		t.Fatalf("first Save() returned error: %v", err)
	}

	err = s.Save(ctx, "https://youtube.com", "AAAAAAAAAA")

	if !errors.Is(err, storage.ErrCodeExists) {
		t.Errorf("expected ErrCodeExists, got %v", err)
	}
}

func TestStorage_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.GetByShort(ctx, "AAAAAAAAAA")

	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
