package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alekshiga/ozon-url-shortener/encoder"
	"github.com/alekshiga/ozon-url-shortener/service"
	"github.com/alekshiga/ozon-url-shortener/storage/memory"
)

func newTestHandler() *Handler {
	store := memory.New()
	gen := encoder.New()
	s := service.New(store, gen)

	return New(s)
}

// проверка функции "укорачивания"
func TestHandler_Shorten(t *testing.T) {
	h := newTestHandler()

	body := []byte(`{"url":"https://google.com"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		bytes.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	var response shortenResponse

	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.ShortURL) != 10 {
		t.Errorf(
			"expected short code length 10, got %d",
			len(response.ShortURL),
		)
	}
}

// проверка GET запроса оригинальной ссылки по коду
func TestHandler_GetOriginal(t *testing.T) {
	h := newTestHandler()

	ctx := context.Background()

	shortCode, err := h.service.Shorten(
		ctx,
		"https://google.com",
	)
	if err != nil {
		t.Fatalf("failed to prepare service: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/urls/"+shortCode,
		nil,
	)

	rec := httptest.NewRecorder()

	h.GetOriginal(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response originalURLResponse

	err = json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.URL != "https://google.com" {
		t.Errorf(
			"expected %q, got %q",
			"https://google.com",
			response.URL,
		)
	}
}

func TestHandler_GetOriginalNotFound(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/urls/OZOOOOOOON",
		nil,
	)

	rec := httptest.NewRecorder()

	h.GetOriginal(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

// проверка пустого POST запроса
func TestHandler_ShortenEmptyURL(t *testing.T) {
	h := newTestHandler()

	body := []byte(`{"url":""}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		bytes.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

// некорректный POST запрос
func TestHandler_ShortenInvalidJSON(t *testing.T) {
	h := newTestHandler()

	body := []byte(`ozon`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		bytes.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
