package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alekshiga/ozon-url-shortener/storage"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connectionString string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	s := &Storage{
		pool: pool,
	}

	if err := s.createTable(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return s, nil
}

func (s *Storage) createTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id BIGSERIAL PRIMARY KEY,
			original_url TEXT NOT NULL UNIQUE,
			short_code VARCHAR(10) NOT NULL UNIQUE
		);
	`

	_, err := s.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("create urls table: %w", err)
	}

	return nil
}

func (s *Storage) Save(ctx context.Context, originalURL string, shortCode string) error {
	query := `
		INSERT INTO urls (original_url, short_code)
		VALUES ($1, $2)
	`

	_, err := s.pool.Exec(ctx, query, originalURL, shortCode)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "urls_original_url_key":
			return storage.ErrURLExists

		case "urls_short_code_key":
			return storage.ErrCodeExists
		}
	}

	return fmt.Errorf("save url: %w", err)
}

func (s *Storage) GetByShort(ctx context.Context, shortCode string) (string, error) {
	query := `
		SELECT original_url
		FROM urls
		WHERE short_code = $1
	`

	var originalURL string

	err := s.pool.QueryRow(ctx, query, shortCode).Scan(&originalURL)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrNotFound
	}

	if err != nil {
		return "", fmt.Errorf("get by short code: %w", err)
	}

	return originalURL, nil
}

func (s *Storage) GetByOriginal(ctx context.Context, originalURL string) (string, error) {
	query := `
		SELECT short_code
		FROM urls
		WHERE original_url = $1
	`

	var shortCode string

	err := s.pool.QueryRow(ctx, query, originalURL).Scan(&shortCode)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrNotFound
	}

	if err != nil {
		return "", fmt.Errorf("get by original URL: %w", err)
	}

	return shortCode, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
