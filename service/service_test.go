package service

import (
	"context"
	"testing"

	"github.com/alekshiga/ozon-url-shortener/encoder"
	"github.com/alekshiga/ozon-url-shortener/storage/memory"
)

func TestService_Shorten(t *testing.T) {
	storage := memory.New()
	coder := encoder.New()
	s := New(storage, coder)

	ctx := context.Background()
	originalURL := "https://google.com"

	shortCode, err := s.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("Shorten() returned error: %v", err)
	}

	if len(shortCode) != 10 {
		t.Errorf("expected short code length 10, got %d", len(shortCode))
	}

	savedURL, err := storage.GetByShort(ctx, shortCode)
	if err != nil {
		t.Fatalf("GetByShort() returned error: %v", err)
	}

	if savedURL != originalURL {
		t.Errorf("expected %q, got %q", originalURL, savedURL)
	}
}

func TestService_ShortenSameURL(t *testing.T) {
	storage := memory.New()
	coder := encoder.New()
	s := New(storage, coder)

	ctx := context.Background()
	originalURL := "https://google.com"

	firstCode, err := s.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("first Shorten() returned error: %v", err)
	}

	secondCode, err := s.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("second Shorten() returned error: %v", err)
	}

	if firstCode != secondCode {
		t.Errorf(
			"expected same short code, got %q and %q",
			firstCode,
			secondCode,
		)
	}
}

func TestService_GetOriginal(t *testing.T) {
	storage := memory.New()
	coder := encoder.New()
	s := New(storage, coder)

	ctx := context.Background()
	originalURL := "https://google.com"

	shortCode, err := s.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("Shorten() returned error: %v", err)
	}

	gotURL, err := s.GetOriginal(ctx, shortCode)
	if err != nil {
		t.Fatalf("GetOriginal() returned error: %v", err)
	}

	if gotURL != originalURL {
		t.Errorf("expected %q, got %q", originalURL, gotURL)
	}
}

type fakeEncoder struct {
	codes []string
	index int
}

func (g *fakeEncoder) Encode() (string, error) {
	code := g.codes[g.index]
	g.index++

	return code, nil
}

func TestService_ShortenRetriesOnCodeCollision(t *testing.T) {
	store := memory.New()
	ctx := context.Background()

	err := store.Save(
		ctx,
		"https://existing.com",
		"AAAAAAAAAA",
	)
	if err != nil {
		t.Fatalf("failed to prepare storage: %v", err)
	}

	gen := &fakeEncoder{
		codes: []string{
			"AAAAAAAAAA",
			"BBBBBBBBBB",
		},
	}

	s := New(store, gen)

	shortCode, err := s.Shorten(
		ctx,
		"https://google.com",
	)
	if err != nil {
		t.Fatalf("Shorten() returned error: %v", err)
	}

	if shortCode != "BBBBBBBBBB" {
		t.Errorf(
			"expected %q, got %q",
			"BBBBBBBBBB",
			shortCode,
		)
	}
}
