package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alekshiga/ozon-url-shortener/service"
	"github.com/alekshiga/ozon-url-shortener/storage"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

type originalURLResponse struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req shortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.URL) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "url is required",
		})
		return
	}

	shortCode, err := h.service.Shorten(
		r.Context(),
		req.URL,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		ShortURL: shortCode,
	})
}

func (h *Handler) GetOriginal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/urls/")

	if shortCode == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "short code is required",
		})
		return
	}

	originalURL, err := h.service.GetOriginal(
		r.Context(),
		shortCode,
	)

	if errors.Is(err, storage.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "url not found",
		})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, originalURLResponse{
		URL: originalURL,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
