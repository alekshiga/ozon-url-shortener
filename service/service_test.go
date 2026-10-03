package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/alekshiga/ozon-url-shortener/encoder"
	"github.com/alekshiga/ozon-url-shortener/storage"
	"github.com/alekshiga/ozon-url-shortener/storage/memory"
)

type fakeEncoder struct {
	codes []string
	index int
}

func (en *fakeEncoder) Encode() (string, error) {
	code := en.codes[en.index]
	en.index++

	return code, nil
}

func TestService_Shorten(t *testing.T) {
	store := memory.New()
	coder := encoder.New()

	service := New(store, coder)

	ctx := context.Background()

	shortCode, err := service.Shorten(
		ctx,
		"https://google.com",
	)
	if err != nil {
		t.Fatalf("Shorten() returned error: %v", err)
	}

	if len(shortCode) != 10 {
		t.Errorf(
			"expected short code length 10, got %d",
			len(shortCode),
		)
	}

	originalURL, err := service.GetOriginal(ctx, shortCode)
	if err != nil {
		t.Fatalf(
			"GetOriginal() returned error: %v",
			err,
		)
	}

	if originalURL != "https://google.com" {
		t.Errorf(
			"expected %q, got %q",
			"https://google.com",
			originalURL,
		)
	}
}

func TestService_SameURLReturnsSameCode(t *testing.T) {
	store := memory.New()
	coder := encoder.New()

	service := New(store, coder)

	ctx := context.Background()
	originalURL := "https://google.com"

	firstCode, err := service.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("first Shorten() returned error: %v", err)
	}

	secondCode, err := service.Shorten(ctx, originalURL)
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

func TestService_CodeCollision(t *testing.T) {
	store := memory.New()

	err := store.Save(
		context.Background(),
		"https://existing.com",
		"AAAAAAAAAA",
	)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	coder := &fakeEncoder{
		codes: []string{
			"AAAAAAAAAA",
			"BBBBBBBBBB",
		},
	}

	service := New(store, coder)

	shortCode, err := service.Shorten(
		context.Background(),
		"https://google.com",
	)
	if err != nil {
		t.Fatalf("Shorten() returned error: %v", err)
	}

	if shortCode != "BBBBBBBBBB" {
		t.Errorf(
			"expected BBBBBBBBBB, got %s",
			shortCode,
		)
	}
}

func TestService_ConcurrentSameURL(t *testing.T) {
	store := memory.New()
	coder := encoder.New()

	service := New(store, coder)

	ctx := context.Background()
	originalURL := "https://google.com"

	const goroutines = 100

	results := make(chan string, goroutines)
	errs := make(chan error, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			shortCode, err := service.Shorten(
				ctx,
				originalURL,
			)

			if err != nil {
				errs <- err
				return
			}

			results <- shortCode
		}()
	}

	wg.Wait()

	close(results)
	close(errs)

	for err := range errs {
		t.Errorf("Shorten() returned error: %v", err)
	}

	var expectedCode string

	for code := range results {
		if expectedCode == "" {
			expectedCode = code
			continue
		}

		if code != expectedCode {
			t.Errorf(
				"expected all goroutines to get %q, got %q",
				expectedCode,
				code,
			)
		}
	}
}

func TestService_ConcurrentDifferentURLs(t *testing.T) {
	store := memory.New()
	coder := encoder.New()

	service := New(store, coder)

	ctx := context.Background()

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errs := make(chan error, goroutines)

	for i := range goroutines {
		go func(i int) {
			defer wg.Done()

			originalURL := fmt.Sprintf(
				"https://example.com/%d",
				i,
			)

			_, err := service.Shorten(
				ctx,
				originalURL,
			)

			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()

	close(errs)

	for err := range errs {
		t.Errorf("Shorten() returned error: %v", err)
	}

	for i := range goroutines {
		originalURL := fmt.Sprintf(
			"https://example.com/%d",
			i,
		)

		_, err := service.Shorten(ctx, originalURL)

		if err != nil {
			t.Errorf(
				"URL %q was not saved correctly: %v",
				originalURL,
				err,
			)
		}
	}
}

func TestService_GetOriginalNotFound(t *testing.T) {
	store := memory.New()
	coder := encoder.New()

	service := New(store, coder)

	_, err := service.GetOriginal(
		context.Background(),
		"AAAAAAAAAA",
	)

	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}
